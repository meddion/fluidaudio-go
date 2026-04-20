package fluidaudio

/*
#cgo CFLAGS: -I${SRCDIR}/include
#cgo LDFLAGS: -L${SRCDIR}/lib/release -lFluidAudioBridge
#cgo LDFLAGS: -L/Library/Developer/CommandLineTools/usr/lib/swift/macosx
#cgo LDFLAGS: -framework Foundation -framework AVFoundation
#cgo LDFLAGS: -framework CoreML -framework Accelerate
#cgo LDFLAGS: -framework Metal -framework MetalPerformanceShaders
#cgo LDFLAGS: -lswiftCore -lc++
#include "fluidaudio.h"
#include <stdlib.h>
*/
import "C"

import (
	"runtime"
	"sync"
	"unsafe"
)

// FluidAudio provides access to FluidAudio's ASR, VAD, Diarization, and Qwen3 capabilities.
type FluidAudio struct {
	ptr  unsafe.Pointer
	mu   sync.Mutex
	done bool
}

// New creates a new FluidAudio bridge instance.
func New() (*FluidAudio, error) {
	ptr := C.fluidaudio_diarizer_create()
	if ptr == nil {
		return nil, &FluidAudioError{Code: -1, Message: "bridge_create returned nil"}
	}
	fa := &FluidAudio{ptr: ptr}
	runtime.SetFinalizer(fa, func(f *FluidAudio) { f.Close() })
	return fa, nil
}

// Close destroys the bridge instance and releases resources. Safe to call multiple times.
func (f *FluidAudio) Close() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.done || f.ptr == nil {
		return
	}
	C.fluidaudio_diarizer_destroy(f.ptr)
	f.ptr = nil
	f.done = true
}

// InitDiarization initializes the speaker diarization engine.
// Pass nil for default config (onsetThreshold=0.5, offsetThreshold=0.5, onsetPadFrames=0).
func (f *FluidAudio) InitDiarization(cfg *DiarizationConfig) error {
	var onsetThreshold float32 = 0.5
	var offsetThreshold float32 = 0.5
	var onsetPadFrames int32 = 0
	compute := All
	variant := VariantDIHARD3
	if cfg != nil {
		if cfg.OnsetThreshold > 0 {
			onsetThreshold = cfg.OnsetThreshold
		}
		if cfg.OffsetThreshold > 0 {
			offsetThreshold = cfg.OffsetThreshold
		}
		onsetPadFrames = cfg.OnsetPadFrames
		compute = cfg.Compute
		variant = cfg.Variant
	}
	rc := C.fluidaudio_initialize_diarization(f.ptr, C.float(onsetThreshold), C.float(offsetThreshold), C.int(onsetPadFrames), C.int(compute), C.int(variant))
	if rc != 0 {
		return &FluidAudioError{Code: int32(rc), Message: "initialize_diarization failed"}
	}
	return nil
}

// DiarizeOffline performs speaker diarization on an audio file.
func (f *FluidAudio) DiarizeOffline(path string) ([]DiarizationSegment, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var speakerIDs *C.int
	var startTimes, endTimes *C.float
	var count C.uint
	rc := C.fluidaudio_diarize_offline(f.ptr, cPath, &speakerIDs, &startTimes, &endTimes, &count)
	if rc != 0 {
		return nil, &FluidAudioError{Code: int32(rc), Message: "fluidaudio_diarize_offline failed"}
	}
	n := int(count)
	if n == 0 {
		return nil, nil
	}
	defer C.fluidaudio_free_segments(speakerIDs, startTimes, endTimes)

	return cSegmentsToGo(speakerIDs, startTimes, endTimes, n), nil
}

// --- Streaming Diarization ---

// ProcessAudio sends a chunk of audio samples to the streaming diarizer.
// sourceSampleRate is the sample rate of the input audio (e.g. 16000).
// Pass 0 to let the diarizer use its native rate.
// Returns only finalized (confirmed) segments. Returns nil if not enough audio has accumulated.
func (f *FluidAudio) ProcessAudio(samples []float32, sourceSampleRate float64) ([]DiarizationSegment, error) {
	if len(samples) == 0 {
		return nil, nil
	}

	var speakerIDs *C.int
	var startTimes, endTimes *C.float
	var count C.uint

	rc := C.fluidaudio_diarize_process_audio(
		f.ptr,
		(*C.float)(unsafe.Pointer(&samples[0])),
		C.uint(len(samples)),
		C.double(sourceSampleRate),
		&speakerIDs, &startTimes, &endTimes, &count,
	)
	if rc != 0 {
		return nil, &FluidAudioError{Code: int32(rc), Message: "diarize_process_audio failed"}
	}

	n := int(count)
	if n == 0 {
		return nil, nil
	}
	defer C.fluidaudio_free_segments(speakerIDs, startTimes, endTimes)

	return cSegmentsToGo(speakerIDs, startTimes, endTimes, n), nil
}

// FinalizeAudio flushes remaining audio through the diarizer, returns all final segments,
// and resets the streaming state. The diarizer remains initialized for a new stream.
func (f *FluidAudio) FinalizeAudio() ([]DiarizationSegment, error) {
	var speakerIDs *C.int
	var startTimes, endTimes *C.float
	var count C.uint

	rc := C.fluidaudio_diarize_finalize(f.ptr, &speakerIDs, &startTimes, &endTimes, &count)
	if rc != 0 {
		return nil, &FluidAudioError{Code: int32(rc), Message: "diarize_finalize failed"}
	}

	n := int(count)
	if n == 0 {
		return nil, nil
	}
	defer C.fluidaudio_free_segments(speakerIDs, startTimes, endTimes)

	return cSegmentsToGo(speakerIDs, startTimes, endTimes, n), nil
}

func cSegmentsToGo(ids *C.int, starts, ends *C.float, n int) []DiarizationSegment {
	if n == 0 {
		return nil
	}
	segments := make([]DiarizationSegment, n)
	cIDs := unsafe.Slice(ids, n)
	cStarts := unsafe.Slice(starts, n)
	cEnds := unsafe.Slice(ends, n)
	for i := 0; i < n; i++ {
		segments[i] = DiarizationSegment{
			SpeakerID: int32(cIDs[i]),
			StartTime: float32(cStarts[i]),
			EndTime:   float32(cEnds[i]),
		}
	}
	return segments
}

// SystemInfo returns platform information.
func (f *FluidAudio) SystemInfo() SystemInfo {
	var cPlatform, cChip *C.char
	C.fluidaudio_get_platform(&cPlatform)
	C.fluidaudio_get_chip_name(&cChip)

	platform := C.GoString(cPlatform)
	chip := C.GoString(cChip)
	C.fluidaudio_free_string(cPlatform)
	C.fluidaudio_free_string(cChip)

	return SystemInfo{
		Platform:     platform,
		ChipName:     chip,
		MemoryGB:     float64(C.fluidaudio_get_memory_gb()),
		AppleSilicon: C.fluidaudio_is_apple_silicon() == 1,
	}
}

// IsAppleSilicon returns whether the current machine has Apple Silicon.
func (f *FluidAudio) IsAppleSilicon() bool {
	return C.fluidaudio_is_apple_silicon() == 1
}
