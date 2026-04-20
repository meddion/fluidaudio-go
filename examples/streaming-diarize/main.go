package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"net/http"
	_ "net/http/pprof" // Registers pprof handlers

	"github.com/meddion/fluidaudio-go"
)

const chunkDurationMs = 500 // Send 500ms chunks to simulate real-time streaming

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <wav-file>\n", os.Args[0])
		os.Exit(1)
	}

	go func() {
		fmt.Println("pprof on https://localhost:6060")
		http.ListenAndServe("localhost:6060", nil)
	}()

	heap, goroutines := metrics()
	fmt.Printf("Before example: heap=%d goroutines=%d\n", heap, goroutines)
	defer func() {
		heapAfter, goroutinesAfter := metrics()
		fmt.Printf("After example: heap=%d (diff=%d) goroutines=%d (diff=%d)\n",
			heapAfter, heapAfter-heap, goroutinesAfter, goroutinesAfter-goroutines)
	}()

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

	// Open WAV file and read header
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer f.Close()

	sampleRate, numChannels, bitsPerSample, err := readWAVHeader(f)
	if err != nil {
		log.Fatalf("Failed to read WAV header: %v", err)
	}
	fmt.Printf("WAV: %d Hz, %d ch, %d bit\n", sampleRate, numChannels, bitsPerSample)

	if bitsPerSample != 16 {
		log.Fatalf("Only 16-bit WAV files are supported, got %d-bit", bitsPerSample)
	}

	// Calculate chunk size in samples
	samplesPerChunk := sampleRate * chunkDurationMs / 1000
	bytesPerSample := int(numChannels) * int(bitsPerSample) / 8
	buf := make([]byte, samplesPerChunk*bytesPerSample)

	fmt.Printf("Streaming %s in %dms chunks...\n", os.Args[1], chunkDurationMs)
	t0 = time.Now()
	chunkNum := 0

	for {
		n, err := io.ReadFull(f, buf)
		if n == 0 {
			break
		}

		// Convert PCM16 to float32, take first channel only if stereo
		samples := pcm16ToFloat32(buf[:n], int(numChannels))

		segments, err2 := fa.ProcessAudio(samples, float64(sampleRate))
		if err2 != nil {
			log.Fatalf("ProcessAudio failed: %v", err2)
		}

		chunkNum++
		chunkTime := float64(chunkNum*chunkDurationMs) / 1000.0

		for _, seg := range segments {
			fmt.Printf("  [%.1fs] Speaker %d: %.2fs - %.2fs\n", chunkTime, seg.SpeakerID, seg.StartTime, seg.EndTime)
		}

		if err == io.ErrUnexpectedEOF || err == io.EOF {
			break
		}
	}

	// Finalize to get all remaining segments
	finalSegments, err := fa.FinalizeAudio()
	if err != nil {
		log.Fatalf("FinalizeAudio failed: %v", err)
	}

	elapsed := time.Since(t0)

	fmt.Printf("\n--- Final Results ---\n")
	fmt.Printf("Total segments: %d\n", len(finalSegments))
	for _, seg := range finalSegments {
		fmt.Printf("  Speaker %d: %.2fs - %.2fs (%.2fs)\n",
			seg.SpeakerID, seg.StartTime, seg.EndTime, seg.EndTime-seg.StartTime)
	}
	fmt.Printf("Streaming diarization took %s\n", elapsed)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

// readWAVHeader parses a WAV file header and returns sample rate, channels, bits per sample.
func readWAVHeader(r io.ReadSeeker) (sampleRate, numChannels, bitsPerSample int, err error) {
	var header struct {
		RiffID   [4]byte
		FileSize uint32
		WaveID   [4]byte
	}
	if err = binary.Read(r, binary.LittleEndian, &header); err != nil {
		return 0, 0, 0, fmt.Errorf("read RIFF header: %w", err)
	}
	if string(header.RiffID[:]) != "RIFF" || string(header.WaveID[:]) != "WAVE" {
		return 0, 0, 0, fmt.Errorf("not a WAV file")
	}

	// Find "fmt " chunk
	for {
		var chunkID [4]byte
		var chunkSize uint32
		if err = binary.Read(r, binary.LittleEndian, &chunkID); err != nil {
			return 0, 0, 0, fmt.Errorf("read chunk ID: %w", err)
		}
		if err = binary.Read(r, binary.LittleEndian, &chunkSize); err != nil {
			return 0, 0, 0, fmt.Errorf("read chunk size: %w", err)
		}

		if string(chunkID[:]) == "fmt " {
			var fmt struct {
				AudioFormat   uint16
				NumChannels   uint16
				SampleRate    uint32
				ByteRate      uint32
				BlockAlign    uint16
				BitsPerSample uint16
			}
			if err = binary.Read(r, binary.LittleEndian, &fmt); err != nil {
				return 0, 0, 0, err
			}
			// Skip any extra fmt bytes
			if chunkSize > 16 {
				r.Seek(int64(chunkSize-16), io.SeekCurrent)
			}
			// Find "data" chunk
			for {
				if err = binary.Read(r, binary.LittleEndian, &chunkID); err != nil {
					return 0, 0, 0, err
				}
				if err = binary.Read(r, binary.LittleEndian, &chunkSize); err != nil {
					return 0, 0, 0, err
				}
				if string(chunkID[:]) == "data" {
					return int(fmt.SampleRate), int(fmt.NumChannels), int(fmt.BitsPerSample), nil
				}
				r.Seek(int64(chunkSize), io.SeekCurrent)
			}
		}
		r.Seek(int64(chunkSize), io.SeekCurrent)
	}
}

// pcm16ToFloat32 converts interleaved PCM16 bytes to mono float32 samples.
func pcm16ToFloat32(data []byte, numChannels int) []float32 {
	numSamples := len(data) / 2 / numChannels
	samples := make([]float32, numSamples)
	for i := 0; i < numSamples; i++ {
		// Take first channel
		offset := i * numChannels * 2
		sample := int16(binary.LittleEndian.Uint16(data[offset : offset+2]))
		samples[i] = float32(sample) / 32768.0
	}
	return samples
}

func metrics() (uint64, int) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	nGoroutines := runtime.NumGoroutine()

	return m.Alloc, nGoroutines
}
