import AVFoundation
import CoreML
import Darwin
import FluidAudio
import Foundation

/// Internal diarizer class that wraps FluidAudio
/// Internal diarization segment used within the diarizer.
enum diarizerError: Error {
    case notInitialized
    case noResult
}

class FluidAudioDiarizer {
    private var lseendDiarizer: LSEENDDiarizer?

    init() {}

    func initializeDiarization(
        compute: MLComputeUnits, variant: LSEENDVariant, cfg: DiarizerTimelineConfig
    ) throws {
        self.lseendDiarizer = LSEENDDiarizer(computeUnits: compute, timelineConfig: cfg)

        let semaphore = DispatchSemaphore(value: 0)
        var initError: Error?
        Task {
            do {
                try await self.lseendDiarizer?.initialize(variant: variant)
            } catch {
                initError = error
            }
            semaphore.signal()
        }
        semaphore.wait()
        if let error = initError {
            throw error
        }
    }

    func diarizeOffline(_ path: String) throws -> [DiarizerSegment] {
        guard let lseendDiarizer = self.lseendDiarizer else {
            throw diarizerError.notInitialized
        }
        let url = URL(fileURLWithPath: path)
        let timeline = try lseendDiarizer.processComplete(audioFileURL: url)

        var segments: [DiarizerSegment] = []
        for (_, speaker) in timeline.speakers {
            segments.append(contentsOf: speaker.finalizedSegments)
        }
        segments.sort { $0.startTime < $1.startTime }

        return segments
    }

    func processAudio(_ audioChunk: [Float32], sourceSample: Double?) throws
        -> DiarizerTimelineUpdate?
    {
        guard let diarizer = self.lseendDiarizer else {
            throw diarizerError.notInitialized
        }

        if let update = try diarizer.process(samples: audioChunk, sourceSampleRate: sourceSample) {
            return update
        }

        return nil
    }

    func finilizeAudio() throws -> DiarizerTimeline {
        guard let diarizer = self.lseendDiarizer else {
            throw diarizerError.notInitialized
        }

        try diarizer.finalizeSession()  // Flush trailing context before reading final output
        let finalTimeline = diarizer.timeline
        diarizer.reset()  // Reset streaming state for a new audio stream (keeps model loaded)

        return finalTimeline
    }

    func cleanup() {
        lseendDiarizer?.cleanup()  // Release all resources including the loaded model
        lseendDiarizer = nil
    }
}

/// Storage for diarizer instances (simple approach - use a single global for now)
private var globalDiarizer: FluidAudioDiarizer?

@_cdecl("fluidaudio_diarizer_create")
public func fluidaudio_diarizer_create() -> UnsafeMutableRawPointer? {
    let diarizer = FluidAudioDiarizer()
    globalDiarizer = diarizer
    return Unmanaged.passRetained(diarizer).toOpaque()
}

@_cdecl("fluidaudio_diarizer_destroy")
public func fluidaudio_diarizer_destroy(_ ptr: UnsafeMutableRawPointer?) {
    guard let ptr = ptr else { return }
    let diarizer = Unmanaged<FluidAudioDiarizer>.fromOpaque(ptr).takeRetainedValue()
    diarizer.cleanup()
    if globalDiarizer === diarizer {
        globalDiarizer = nil
    }
}

@_cdecl("fluidaudio_diarize_offline")
public func fluidaudio_diarize_offline(
    _ diarizer_ptr: UnsafeMutableRawPointer,
    _ path: UnsafePointer<CChar>,
    // Segment data
    _ outSpeakerIds: UnsafeMutablePointer<UnsafeMutablePointer<Int32>>,
    _ outStartTimes: UnsafeMutablePointer<UnsafeMutablePointer<Float>>,
    _ outEndTimes: UnsafeMutablePointer<UnsafeMutablePointer<Float>>,
    _ outCount: UnsafeMutablePointer<UInt32>
) -> Int32 {
    let pathString = String(cString: path)
    let diarizer = Unmanaged<FluidAudioDiarizer>.fromOpaque(diarizer_ptr).takeUnretainedValue()

    do {
        let segments = try diarizer.diarizeOffline(pathString)
        let count = segments.count
        outCount.pointee = UInt32(count)

        if count > 0 {
            let ids = UnsafeMutablePointer<Int32>.allocate(capacity: count)
            let starts = UnsafeMutablePointer<Float>.allocate(capacity: count)
            let ends = UnsafeMutablePointer<Float>.allocate(capacity: count)

            for (i, seg) in segments.enumerated() {
                ids[i] = Int32(seg.speakerIndex)
                starts[i] = seg.startTime
                ends[i] = seg.endTime
            }

            outSpeakerIds.pointee = ids
            outStartTimes.pointee = starts
            outEndTimes.pointee = ends
        }

        return 0
    } catch {
        print("Diarize error: \(error)")
        return -1
    }
}

@_cdecl("fluidaudio_initialize_diarization")
public func fluidaudio_initialize_diarization(
    _ diarizer_ptr: UnsafeMutableRawPointer,
    onset_threshold: Float32,
    offset_threshold: Float32,
    onset_pad_frames: Int32,
    offset_pad_frames: Int32,
    min_frames_on: Int32,
    min_frames_off: Int32,
    compute_type: Int32,
    variant_type: Int32
) -> Int32 {
    let diarizer = Unmanaged<FluidAudioDiarizer>.fromOpaque(diarizer_ptr).takeUnretainedValue()

    let compute = MLComputeUnits(rawValue: Int(compute_type)) ?? .all

    let variants: [LSEENDVariant] = [.ami, .callhome, .dihard2, .dihard3]
    let variant =
        variants.indices.contains(Int(variant_type)) ? variants[Int(variant_type)] : .dihard3

    do {
        let cfg = DiarizerTimelineConfig(
            onsetThreshold: onset_threshold,
            offsetThreshold: offset_threshold,
            onsetPadFrames: Int(onset_pad_frames),
            offsetPadFrames: Int(offset_pad_frames),
            minFramesOn: Int(min_frames_on),
            minFramesOff: Int(min_frames_off)
        )
        try diarizer.initializeDiarization(compute: compute, variant: variant, cfg: cfg)
        return 0
    } catch {
        print("Diarize error: \(error)")
        return -1
    }
}

/// Helper to write segment arrays into pre-allocated C output pointers.
private func writeSegments(
    _ segments: [DiarizerSegment],
    ids: UnsafeMutablePointer<UnsafeMutablePointer<Int32>>,
    starts: UnsafeMutablePointer<UnsafeMutablePointer<Float>>,
    ends: UnsafeMutablePointer<UnsafeMutablePointer<Float>>,
    count: UnsafeMutablePointer<UInt32>
) {
    let n = segments.count
    count.pointee = UInt32(n)
    if n > 0 {
        let idsPtr = UnsafeMutablePointer<Int32>.allocate(capacity: n)
        let startsPtr = UnsafeMutablePointer<Float>.allocate(capacity: n)
        let endsPtr = UnsafeMutablePointer<Float>.allocate(capacity: n)
        for (i, seg) in segments.enumerated() {
            idsPtr[i] = Int32(seg.speakerIndex)
            startsPtr[i] = seg.startTime
            endsPtr[i] = seg.endTime
        }
        ids.pointee = idsPtr
        starts.pointee = startsPtr
        ends.pointee = endsPtr
    }
}

@_cdecl("fluidaudio_diarize_process_audio")
public func fluidaudio_diarize_process_audio(
    _ diarizer_ptr: UnsafeMutableRawPointer,
    _ samples: UnsafePointer<Float>,
    _ sampleCount: UInt32,
    _ sourceSampleRate: Double,
    _ outSpeakerIds: UnsafeMutablePointer<UnsafeMutablePointer<Int32>>,
    _ outStartTimes: UnsafeMutablePointer<UnsafeMutablePointer<Float>>,
    _ outEndTimes: UnsafeMutablePointer<UnsafeMutablePointer<Float>>,
    _ outCount: UnsafeMutablePointer<UInt32>
) -> Int32 {
    let diarizer = Unmanaged<FluidAudioDiarizer>.fromOpaque(diarizer_ptr).takeUnretainedValue()
    let chunk = Array(UnsafeBufferPointer(start: samples, count: Int(sampleCount)))
    let rate: Double? = sourceSampleRate > 0 ? sourceSampleRate : nil

    do {
        let update = try diarizer.processAudio(chunk, sourceSample: rate)
        if let update = update {
            writeSegments(
                update.finalizedSegments,
                ids: outSpeakerIds, starts: outStartTimes, ends: outEndTimes,
                count: outCount)
        } else {
            outCount.pointee = 0
        }
        return 0
    } catch {
        print("Streaming diarize error: \(error)")
        return -1
    }
}

@_cdecl("fluidaudio_diarize_finalize")
public func fluidaudio_diarize_finalize(
    _ diarizer_ptr: UnsafeMutableRawPointer,
    _ outSpeakerIds: UnsafeMutablePointer<UnsafeMutablePointer<Int32>>,
    _ outStartTimes: UnsafeMutablePointer<UnsafeMutablePointer<Float>>,
    _ outEndTimes: UnsafeMutablePointer<UnsafeMutablePointer<Float>>,
    _ outCount: UnsafeMutablePointer<UInt32>
) -> Int32 {
    let diarizer = Unmanaged<FluidAudioDiarizer>.fromOpaque(diarizer_ptr).takeUnretainedValue()

    do {
        let timeline = try diarizer.finilizeAudio()
        var segments: [DiarizerSegment] = []
        for (_, speaker) in timeline.speakers {
            segments.append(contentsOf: speaker.finalizedSegments)
        }
        segments.sort { $0.startTime < $1.startTime }
        writeSegments(
            segments,
            ids: outSpeakerIds, starts: outStartTimes, ends: outEndTimes,
            count: outCount)
        return 0
    } catch {
        print("Finalize diarize error: \(error)")
        return -1
    }
}

@_cdecl("fluidaudio_free_segments")
public func fluidaudio_free_segments(
    _ speakerIds: UnsafeMutablePointer<Int32>?,
    _ startTimes: UnsafeMutablePointer<Float>?,
    _ endTimes: UnsafeMutablePointer<Float>?
) {
    speakerIds?.deallocate()
    startTimes?.deallocate()
    endTimes?.deallocate()
}

@_cdecl("fluidaudio_get_platform")
public func fluidaudio_get_platform(_ out: UnsafeMutablePointer<UnsafeMutablePointer<CChar>?>?) {
    #if os(macOS)
        let platform = "macOS"
    #elseif os(iOS)
        let platform = "iOS"
    #else
        let platform = "unknown"
    #endif

    out?.pointee = strdup(platform)
}

@_cdecl("fluidaudio_get_chip_name")
public func fluidaudio_get_chip_name(_ out: UnsafeMutablePointer<UnsafeMutablePointer<CChar>?>?) {
    var size: size_t = 0
    var chipName = "Unknown"

    if sysctlbyname("machdep.cpu.brand_string", nil, &size, nil, 0) == 0, size > 0 {
        var buffer = [CChar](repeating: 0, count: Int(size))
        if sysctlbyname("machdep.cpu.brand_string", &buffer, &size, nil, 0) == 0 {
            chipName = String(cString: buffer)
        }
    }

    out?.pointee = strdup(chipName)
}

@_cdecl("fluidaudio_get_memory_gb")
public func fluidaudio_get_memory_gb() -> Double {
    return Double(ProcessInfo.processInfo.physicalMemory) / (1024 * 1024 * 1024)
}

@_cdecl("fluidaudio_is_apple_silicon")
public func fluidaudio_is_apple_silicon() -> Int32 {
    return SystemInfo.isAppleSilicon ? 1 : 0
}

@_cdecl("fluidaudio_cleanup")
public func fluidaudio_cleanup(_ ptr: UnsafeMutableRawPointer?) {
    guard let ptr = ptr else { return }
    let diarizer = Unmanaged<FluidAudioDiarizer>.fromOpaque(ptr).takeUnretainedValue()
    diarizer.cleanup()
}

@_cdecl("fluidaudio_free_string")
public func fluidaudio_free_string(_ s: UnsafeMutablePointer<CChar>?) {
    free(s)
}
