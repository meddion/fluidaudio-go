package fluidaudio

import "fmt"

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

// DiarizationVariant selects the pre-trained diarization model variant.
type DiarizationVariant int32

const (
	VariantAMI      DiarizationVariant = 0
	VariantCallhome DiarizationVariant = 1
	VariantDIHARD2  DiarizationVariant = 2
	VariantDIHARD3  DiarizationVariant = 3 // default
)

// DiarizationSegment represents a single speaker segment.
type DiarizationSegment struct {
	SpeakerID int32
	StartTime float32
	EndTime   float32
}

// SystemInfo contains platform information.
type SystemInfo struct {
	Platform     string
	ChipName     string
	MemoryGB     float64
	AppleSilicon bool
}

// FluidAudioError represents an error from the FluidAudio bridge.
type FluidAudioError struct {
	Code    int32
	Message string
}

func (e *FluidAudioError) Error() string {
	return fmt.Sprintf("fluidaudio: %s (code %d)", e.Message, e.Code)
}
