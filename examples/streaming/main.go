package main

import (
	"fmt"
	"log"
	"os"

	"github.com/meddion/fluidaudio-go"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <audio-file>\n", os.Args[0])
		os.Exit(1)
	}

	fa, err := fluidaudio.New()
	if err != nil {
		log.Fatalf("Failed to create FluidAudio: %v", err)
	}
	defer fa.Close()

	fmt.Println("Initializing Streaming ASR...")
	if err := fa.InitStreamingASR(); err != nil {
		log.Fatalf("Failed to init Streaming ASR: %v", err)
	}

	fmt.Printf("Transcribing %s (streaming)...\n", os.Args[1])
	result, err := fa.TranscribeFileStreaming(os.Args[1])
	if err != nil {
		log.Fatalf("Streaming transcription failed: %v", err)
	}

	fmt.Printf("Text: %s\n", result.Text)
	fmt.Printf("Confidence: %.2f\n", result.Confidence)
	fmt.Printf("Duration: %.2fs\n", result.Duration)
	fmt.Printf("Processing time: %.2fs\n", result.ProcessingTime)
	fmt.Printf("RTFx: %.2f\n", result.RTFx)
}
