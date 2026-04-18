package fluidaudio

import "fmt"

// AsrResult contains the result of a speech recognition operation.
type AsrResult struct {
	Text           string
	Confidence     float32
	Duration       float64
	ProcessingTime float64
	RTFx           float32
}

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
