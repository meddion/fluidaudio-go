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

	info := fa.SystemInfo()
	fmt.Printf("Platform: %s, Chip: %s, Memory: %.1f GB\n", info.Platform, info.ChipName, info.MemoryGB)

	fmt.Println("Initializing ASR...")
	if err := fa.InitASR(); err != nil {
		log.Fatalf("Failed to init ASR: %v", err)
	}

	fmt.Printf("Transcribing %s...\n", os.Args[1])
	result, err := fa.TranscribeFile(os.Args[1])
	if err != nil {
		log.Fatalf("Transcription failed: %v", err)
	}

	fmt.Printf("Text: %s\n", result.Text)
	fmt.Printf("Confidence: %.2f\n", result.Confidence)
	fmt.Printf("Duration: %.2fs\n", result.Duration)
	fmt.Printf("Processing time: %.2fs\n", result.ProcessingTime)
	fmt.Printf("RTFx: %.2f\n", result.RTFx)
}
