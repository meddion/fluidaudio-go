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

// --- Diarization ---

// DiarizationConfig holds configuration for speaker diarization.
type DiarizationConfig struct {
	// OnsetThreshold is the probability threshold to start a speech segment (default 0.5).
	// Higher = fewer false-positive speech onsets.
	OnsetThreshold float32
	// OffsetThreshold is the probability threshold to end a speech segment (default 0.5).
	// Lower = segments sustained longer through probability dips.
	OffsetThreshold float32
	// OnsetPadFrames is the number of frames to pad before each speech onset (default 0).
	OnsetPadFrames int32
	Compute        ComputeType
	// Variant selects the pre-trained model variant (default VariantDIHARD3).
	Variant DiarizationVariant
}

type ComputeType uint8

const (
	CPUOnly ComputeType = iota
	CPUAndNeuralEngine
	CPUAndGPU
	All
)

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
	defer C.fluidaudio_free_diarize_offline(speakerIDs, startTimes, endTimes)

	n := int(count)
	if n == 0 {
		return nil, nil
	}
	segments := make([]DiarizationSegment, n)
	ids := unsafe.Slice(speakerIDs, n)
	starts := unsafe.Slice(startTimes, n)
	ends := unsafe.Slice(endTimes, n)
	for i := 0; i < n; i++ {
		segments[i] = DiarizationSegment{
			SpeakerID: int32(ids[i]),
			StartTime: float32(starts[i]),
			EndTime:   float32(ends[i]),
		}
	}

	return segments, nil
}

/// --- System Info ---

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
