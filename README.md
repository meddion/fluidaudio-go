# fluidaudio-go

Go bindings for [FluidAudio](https://github.com/FluidInference/FluidAudio). Right now supports only speaker diarization on Apple platforms. Uses LS-EEND (Linear Streaming End-to-End Neural Diarization) via CoreML.

## Features

- **Offline diarization** — process a complete audio file
- **Streaming diarization** — feed audio chunks incrementally, get finalized segments as they're confirmed
- **Configurable** — onset/offset thresholds, onset padding, compute units (CPU/GPU/Neural Engine), model variants (AMI, CALLHOME, DIHARD II/III)
- **System info** — platform, chip name, memory, Apple Silicon detection

## Requirements

- macOS 14+
- Apple Silicon recommended
- Swift 5.10+
- Go 1.21+

## Build

```bash
# Build Swift static library + Go package
make

# Run tests
make test

# Run integration tests (requires model downloads)
make test-integration
```

## Usage

### Offline diarization

```go
fa, _ := fluidaudio.New()
defer fa.Close()

fa.InitDiarization(nil) // default config

segments, _ := fa.DiarizeOffline("meeting.wav")
for _, seg := range segments {
    fmt.Printf("Speaker %d: %.2fs - %.2fs\n", seg.SpeakerID, seg.StartTime, seg.EndTime)
}
```

### Streaming diarization

```go
fa, _ := fluidaudio.New()
defer fa.Close()

fa.InitDiarization(nil)

// Feed audio chunks as they arrive
for chunk := range audioChunks {
    segments, _ := fa.ProcessAudio(chunk, 16000)
    for _, seg := range segments {
        fmt.Printf("Speaker %d: %.2fs - %.2fs\n", seg.SpeakerID, seg.StartTime, seg.EndTime)
    }
}

// Flush remaining audio and get all final segments
finalSegments, _ := fa.FinalizeAudio()
```

### Configuration

```go
fa.InitDiarization(&fluidaudio.DiarizationConfig{
    OnsetThreshold:  0.5,              // probability to start a speech segment
    OffsetThreshold: 0.5,              // probability to end a speech segment
    OnsetPadFrames:  0,                // frames to pad before speech onset
    Compute:         fluidaudio.All,   // CPU, CPUAndGPU, CPUAndNeuralEngine, All
    Variant:         fluidaudio.VariantAMI, // AMI, CALLHOME, DIHARD2, DIHARD3
})
```

## Examples

```bash
# Offline diarization
go run ./examples/diarize/ audio.wav

# Streaming diarization (simulated from file)
go run ./examples/streaming-diarize/ audio.wav
```

## License

See [FluidAudio](https://github.com/FluidInference/FluidAudio) for licensing terms.
