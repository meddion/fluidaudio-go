# fluidaudio-go — Agent Instructions

Go bindings for FluidAudio, providing speaker diarization on macOS via CoreML. Uses CGO to bridge Go → C → Swift.

## Build & Test

```bash
make              # Build Swift static lib + Go package
make test         # Unit tests (go test -v -count=1)
make test-integration  # Integration tests (downloads models, slow)
make clean        # Remove build artifacts
```

**Requirements**: macOS 14+, Apple Silicon, Swift 5.10+, Go 1.26+

## Public API

- `New() → *FluidAudio` — Create bridge instance
- `Close()` — Cleanup (idempotent, also registered as runtime finalizer)
- `InitDiarization(cfg)` — Load diarization model (variants: AMI, CALLHOME, DIHARD2, DIHARD3)
- `ProcessAudio(samples, sampleRate) → []DiarizationSegment` — Stream audio chunks
- `FinalizeAudio() → []DiarizationSegment` — Flush remaining audio, reset stream
- `DiarizeOffline(path) → []DiarizationSegment` — Process complete audio file
- `SystemInfo()` — Platform, chip, memory info

## Architecture

```
Go (fluidaudio.go) → CGO → C header (include/fluidaudio.h) → Swift (swift/FluidAudioBridge.swift) → FluidAudio framework
```

- CGO preamble in [fluidaudio.go](fluidaudio.go) links against Swift runtime, CoreML, Metal, Accelerate
- Types defined in [fluidaudio_types.go](fluidaudio_types.go)
- Swift library built as static `.a` via `swift build -c release`

## CGO Patterns

- Input strings: `C.CString()` + `defer C.free()`
- Output strings: `C.GoString()` + `C.free()`
- Output arrays: `unsafe.Slice()` then `C.fluidaudio_free_segments()`
- All C memory must be explicitly freed
