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

// DiarizationSegment represents a single speaker segment.
type DiarizationSegment struct {
	SpeakerID    string
	StartTime    float32
	EndTime      float32
	QualityScore float32
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
