# fluidaudio-go

Go bindings for [FluidAudio](https://github.com/FluidInference/FluidAudio) — ASR, VAD, Speaker Diarization, and Qwen3 on Apple platforms.

## Requirements

- macOS 14+ (macOS 15+ for Qwen3)
- Apple Silicon recommended
- Swift 5.10+
- Go 1.21+

## Build

```bash
# Build the Swift static library (one-time, or after Swift changes)
make

# Build Go package
go build ./...

# Run tests
go test ./...

# Run integration tests (requires model downloads)
go test -tags integration ./...
```

## Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/meddion/fluidaudio-go"
)

func main() {
	fa, err := fluidaudio.New()
	if err != nil {
		log.Fatal(err)
	}
	defer fa.Close()

	// Print system info
	info := fa.SystemInfo()
	fmt.Printf("Running on %s (%s)\n", info.Platform, info.ChipName)

	// Initialize and transcribe
	if err := fa.InitASR(); err != nil {
		log.Fatal(err)
	}

	result, err := fa.TranscribeFile("audio.wav")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Text: %s (confidence: %.2f)\n", result.Text, result.Confidence)
}
```

## Features

- **ASR** — File and sample-based transcription (Parakeet TDT, English-optimized)
- **Streaming ASR** — Low-memory streaming transcription
- **VAD** — Voice Activity Detection
- **Speaker Diarization** — Identify who spoke when
- **Qwen3 ASR** — Multilingual transcription (30+ languages)
- **Qwen3 Streaming** — Real-time multilingual streaming
- **System Info** — Platform, chip, memory queries

## Examples

```bash
go run ./examples/transcribe/ audio.wav
go run ./examples/streaming/ audio.wav
go run ./examples/diarize/ audio.wav
```

## License

See [FluidAudio](https://github.com/FluidInference/FluidAudio) for licensing terms.
