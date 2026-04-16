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

	fmt.Println("Initializing Diarization...")
	cfg := fluidaudio.DiarizationConfig{
		Threshold:   0.7,
		MinSpeakers: 2,
		MaxSpeakers: 5,
	}
	if err := fa.InitDiarization(&cfg); err != nil {
		log.Fatalf("Failed to init Diarization: %v", err)
	}

	fmt.Printf("Diarizing %s...\n", os.Args[1])
	segments, err := fa.DiarizeFile(os.Args[1])
	if err != nil {
		log.Fatalf("Diarization failed: %v", err)
	}

	fmt.Printf("Found %d segments:\n", len(segments))
	for _, seg := range segments {
		fmt.Printf("  [%s] %.2fs - %.2fs (quality: %.2f)\n",
			seg.SpeakerID, seg.StartTime, seg.EndTime, seg.QualityScore)
	}
}
