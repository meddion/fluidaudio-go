package main

import (
	"fmt"
	"log"
	"os"
	"time"

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

	fmt.Printf("System info: %+v\n", fa.SystemInfo())

	t0 := time.Now()
	if err := fa.InitDiarization(nil); err != nil {
		log.Fatalf("Diarization failed to init: %v", err)
	}
	fmt.Printf("Diarizer init took %s\n", time.Since(t0))

	fmt.Printf("Diarizing %s...\n", os.Args[1])

	t0 = time.Now()
	segments, err := fa.DiarizeOffline(os.Args[1])
	if err != nil {
		log.Fatalf("Diarization failed: %v", err)
	}

	fmt.Printf("Found %d segments:\n", len(segments))
	for _, seg := range segments {
		fmt.Printf("  [%d] %.2fs - %.2fs\n",
			seg.SpeakerID, seg.StartTime, seg.EndTime)
	}
	fmt.Printf("Diarization took %s\n", time.Since(t0))
}
