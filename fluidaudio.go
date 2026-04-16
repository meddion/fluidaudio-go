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
	ptr := C.fluidaudio_bridge_create()
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
	C.fluidaudio_bridge_destroy(f.ptr)
	f.ptr = nil
	f.done = true
}

// --- ASR ---

// InitASR initializes the ASR engine. Must be called before transcription.
func (f *FluidAudio) InitASR() error {
	rc := C.fluidaudio_initialize_asr(f.ptr)
	if rc != 0 {
		return &FluidAudioError{Code: int32(rc), Message: "initialize_asr failed"}
	}
	return nil
}

// TranscribeFile transcribes speech from an audio file.
func (f *FluidAudio) TranscribeFile(path string) (*AsrResult, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var cText *C.char
	var confidence C.float
	var duration, processingTime C.double
	var rtfx C.float

	rc := C.fluidaudio_transcribe_file(f.ptr, cPath,
		&cText, &confidence, &duration, &processingTime, &rtfx)
	if rc != 0 {
		return nil, &FluidAudioError{Code: int32(rc), Message: "transcribe_file failed"}
	}

	text := C.GoString(cText)
	C.fluidaudio_free_string(cText)

	return &AsrResult{
		Text:           text,
		Confidence:     float32(confidence),
		Duration:       float64(duration),
		ProcessingTime: float64(processingTime),
		RTFx:           float32(rtfx),
	}, nil
}

// TranscribeSamples transcribes speech from raw float32 audio samples (16kHz mono).
func (f *FluidAudio) TranscribeSamples(samples []float32) (*AsrResult, error) {
	if len(samples) == 0 {
		return nil, &FluidAudioError{Code: -1, Message: "empty samples"}
	}

	var cText *C.char
	var confidence C.float
	var duration, processingTime C.double
	var rtfx C.float

	rc := C.fluidaudio_transcribe_samples(f.ptr,
		(*C.float)(unsafe.Pointer(&samples[0])),
		C.uint(len(samples)),
		&cText, &confidence, &duration, &processingTime, &rtfx)
	if rc != 0 {
		return nil, &FluidAudioError{Code: int32(rc), Message: "transcribe_samples failed"}
	}

	text := C.GoString(cText)
	C.fluidaudio_free_string(cText)

	return &AsrResult{
		Text:           text,
		Confidence:     float32(confidence),
		Duration:       float64(duration),
		ProcessingTime: float64(processingTime),
		RTFx:           float32(rtfx),
	}, nil
}

// IsASRAvailable returns whether the ASR engine is initialized and ready.
func (f *FluidAudio) IsASRAvailable() bool {
	return C.fluidaudio_is_asr_available(f.ptr) == 1
}

// --- Streaming ASR ---

// InitStreamingASR initializes the streaming ASR engine.
func (f *FluidAudio) InitStreamingASR() error {
	rc := C.fluidaudio_initialize_streaming_asr(f.ptr)
	if rc != 0 {
		return &FluidAudioError{Code: int32(rc), Message: "initialize_streaming_asr failed"}
	}
	return nil
}

// StreamingASRStart begins a new streaming ASR session.
func (f *FluidAudio) StreamingASRStart() error {
	rc := C.fluidaudio_streaming_asr_start(f.ptr)
	if rc != 0 {
		return &FluidAudioError{Code: int32(rc), Message: "streaming_asr_start failed"}
	}
	return nil
}

// StreamingASRFeed feeds audio samples into the streaming ASR session.
func (f *FluidAudio) StreamingASRFeed(samples []float32) error {
	if len(samples) == 0 {
		return nil
	}
	rc := C.fluidaudio_streaming_asr_feed(f.ptr,
		(*C.float)(unsafe.Pointer(&samples[0])),
		C.uint(len(samples)))
	if rc != 0 {
		return &FluidAudioError{Code: int32(rc), Message: "streaming_asr_feed failed"}
	}
	return nil
}

// StreamingASRFinish completes the streaming ASR session and returns the final text.
func (f *FluidAudio) StreamingASRFinish() (string, error) {
	var cText *C.char
	rc := C.fluidaudio_streaming_asr_finish(f.ptr, &cText)
	if rc != 0 {
		return "", &FluidAudioError{Code: int32(rc), Message: "streaming_asr_finish failed"}
	}
	text := C.GoString(cText)
	C.fluidaudio_free_string(cText)
	return text, nil
}

// TranscribeFileStreaming transcribes a file using the streaming engine.
func (f *FluidAudio) TranscribeFileStreaming(path string) (*AsrResult, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var cText *C.char
	var confidence C.float
	var duration, processingTime C.double
	var rtfx C.float

	rc := C.fluidaudio_transcribe_file_streaming(f.ptr, cPath,
		&cText, &confidence, &duration, &processingTime, &rtfx)
	if rc != 0 {
		return nil, &FluidAudioError{Code: int32(rc), Message: "transcribe_file_streaming failed"}
	}

	text := C.GoString(cText)
	C.fluidaudio_free_string(cText)

	return &AsrResult{
		Text:           text,
		Confidence:     float32(confidence),
		Duration:       float64(duration),
		ProcessingTime: float64(processingTime),
		RTFx:           float32(rtfx),
	}, nil
}

// IsStreamingASRAvailable returns whether the streaming ASR engine is initialized.
func (f *FluidAudio) IsStreamingASRAvailable() bool {
	return C.fluidaudio_is_streaming_asr_available(f.ptr) == 1
}

// --- VAD ---

// InitVAD initializes the Voice Activity Detection engine.
func (f *FluidAudio) InitVAD(threshold float32) error {
	rc := C.fluidaudio_initialize_vad(f.ptr, C.float(threshold))
	if rc != 0 {
		return &FluidAudioError{Code: int32(rc), Message: "initialize_vad failed"}
	}
	return nil
}

// IsVADAvailable returns whether the VAD engine is initialized.
func (f *FluidAudio) IsVADAvailable() bool {
	return C.fluidaudio_is_vad_available(f.ptr) == 1
}

// --- Diarization ---

// DiarizationConfig holds configuration for speaker diarization.
type DiarizationConfig struct {
	// Threshold is the clustering distance threshold (default 0.6).
	// Higher = fewer speakers, lower = more speakers. Range: (0, √2].
	Threshold float64
	// MinSpeakers sets the minimum number of speakers. 0 means no constraint.
	MinSpeakers int
	// MaxSpeakers sets the maximum number of speakers. 0 means no constraint.
	MaxSpeakers int
}

// InitDiarization initializes the speaker diarization engine.
// Pass nil for default config (threshold=0.6, no speaker constraints).
func (f *FluidAudio) InitDiarization(cfg *DiarizationConfig) error {
	threshold := 0.6
	var minSpeakers, maxSpeakers C.int32_t
	if cfg != nil {
		if cfg.Threshold > 0 {
			threshold = cfg.Threshold
		}
		minSpeakers = C.int32_t(cfg.MinSpeakers)
		maxSpeakers = C.int32_t(cfg.MaxSpeakers)
	}
	rc := C.fluidaudio_initialize_diarization(f.ptr, C.double(threshold), minSpeakers, maxSpeakers)
	if rc != 0 {
		return &FluidAudioError{Code: int32(rc), Message: "initialize_diarization failed"}
	}
	return nil
}

// DiarizeFile performs speaker diarization on an audio file.
func (f *FluidAudio) DiarizeFile(path string) ([]DiarizationSegment, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var speakerIds **C.char
	var startTimes, endTimes, qualityScores *C.float
	var count C.uint

	rc := C.fluidaudio_diarize_file(f.ptr, cPath,
		&speakerIds, &startTimes, &endTimes, &qualityScores, &count)
	if rc != 0 {
		return nil, &FluidAudioError{Code: int32(rc), Message: "diarize_file failed"}
	}
	defer C.fluidaudio_free_diarization_result(speakerIds, startTimes, endTimes, qualityScores, count)

	n := int(count)
	if n == 0 {
		return nil, nil
	}

	segments := make([]DiarizationSegment, n)
	ids := unsafe.Slice(speakerIds, n)
	starts := unsafe.Slice(startTimes, n)
	ends := unsafe.Slice(endTimes, n)
	scores := unsafe.Slice(qualityScores, n)

	for i := 0; i < n; i++ {
		segments[i] = DiarizationSegment{
			SpeakerID:    C.GoString(ids[i]),
			StartTime:    float32(starts[i]),
			EndTime:      float32(ends[i]),
			QualityScore: float32(scores[i]),
		}
	}
	return segments, nil
}

// IsDiarizationAvailable returns whether diarization is initialized.
func (f *FluidAudio) IsDiarizationAvailable() bool {
	return C.fluidaudio_is_diarization_available(f.ptr) == 1
}

// --- Qwen3 ASR ---

// InitQwen3ASR initializes the Qwen3 ASR engine. Requires macOS 15+ or iOS 18+.
func (f *FluidAudio) InitQwen3ASR() error {
	rc := C.fluidaudio_initialize_qwen3_asr(f.ptr)
	if rc != 0 {
		return &FluidAudioError{Code: int32(rc), Message: "initialize_qwen3_asr failed"}
	}
	return nil
}

// Qwen3TranscribeFile transcribes a file using the Qwen3 model.
// Pass nil for language to use automatic detection.
func (f *FluidAudio) Qwen3TranscribeFile(path string, language *string) (*AsrResult, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var cLang *C.char
	if language != nil {
		cLang = C.CString(*language)
		defer C.free(unsafe.Pointer(cLang))
	}

	var cText *C.char
	var confidence C.float
	var duration, processingTime C.double
	var rtfx C.float

	rc := C.fluidaudio_qwen3_transcribe_file(f.ptr, cPath, cLang,
		&cText, &confidence, &duration, &processingTime, &rtfx)
	if rc != 0 {
		return nil, &FluidAudioError{Code: int32(rc), Message: "qwen3_transcribe_file failed"}
	}

	text := C.GoString(cText)
	C.fluidaudio_free_string(cText)

	return &AsrResult{
		Text:           text,
		Confidence:     float32(confidence),
		Duration:       float64(duration),
		ProcessingTime: float64(processingTime),
		RTFx:           float32(rtfx),
	}, nil
}

// Qwen3TranscribeSamples transcribes raw audio samples using the Qwen3 model.
// Pass nil for language to use automatic detection.
func (f *FluidAudio) Qwen3TranscribeSamples(samples []float32, language *string) (*AsrResult, error) {
	if len(samples) == 0 {
		return nil, &FluidAudioError{Code: -1, Message: "empty samples"}
	}

	var cLang *C.char
	if language != nil {
		cLang = C.CString(*language)
		defer C.free(unsafe.Pointer(cLang))
	}

	var cText *C.char
	var confidence C.float
	var duration, processingTime C.double
	var rtfx C.float

	rc := C.fluidaudio_qwen3_transcribe_samples(f.ptr,
		(*C.float)(unsafe.Pointer(&samples[0])),
		C.uint(len(samples)),
		cLang, &cText, &confidence, &duration, &processingTime, &rtfx)
	if rc != 0 {
		return nil, &FluidAudioError{Code: int32(rc), Message: "qwen3_transcribe_samples failed"}
	}

	text := C.GoString(cText)
	C.fluidaudio_free_string(cText)

	return &AsrResult{
		Text:           text,
		Confidence:     float32(confidence),
		Duration:       float64(duration),
		ProcessingTime: float64(processingTime),
		RTFx:           float32(rtfx),
	}, nil
}

// IsQwen3ASRAvailable returns whether the Qwen3 ASR engine is initialized.
func (f *FluidAudio) IsQwen3ASRAvailable() bool {
	return C.fluidaudio_is_qwen3_asr_available(f.ptr) == 1
}

// --- Qwen3 Streaming ---

// InitQwen3Streaming initializes the Qwen3 streaming engine.
func (f *FluidAudio) InitQwen3Streaming() error {
	rc := C.fluidaudio_initialize_qwen3_streaming(f.ptr)
	if rc != 0 {
		return &FluidAudioError{Code: int32(rc), Message: "initialize_qwen3_streaming failed"}
	}
	return nil
}

// Qwen3StreamingStart begins a new Qwen3 streaming session.
// Pass nil for language to use automatic detection.
func (f *FluidAudio) Qwen3StreamingStart(language *string, minAudioSec, chunkSec, maxAudioSec float64) error {
	var cLang *C.char
	if language != nil {
		cLang = C.CString(*language)
		defer C.free(unsafe.Pointer(cLang))
	}

	rc := C.fluidaudio_qwen3_streaming_start(f.ptr, cLang,
		C.double(minAudioSec), C.double(chunkSec), C.double(maxAudioSec))
	if rc != 0 {
		return &FluidAudioError{Code: int32(rc), Message: "qwen3_streaming_start failed"}
	}
	return nil
}

// Qwen3StreamingFeed feeds audio samples into the Qwen3 streaming session.
// Returns partial transcription text if available, or nil if no partial result yet.
func (f *FluidAudio) Qwen3StreamingFeed(samples []float32) (*string, error) {
	if len(samples) == 0 {
		return nil, nil
	}

	var cPartial *C.char
	rc := C.fluidaudio_qwen3_streaming_feed(f.ptr,
		(*C.float)(unsafe.Pointer(&samples[0])),
		C.uint(len(samples)),
		&cPartial)
	if rc != 0 {
		return nil, &FluidAudioError{Code: int32(rc), Message: "qwen3_streaming_feed failed"}
	}

	if cPartial == nil {
		return nil, nil
	}

	text := C.GoString(cPartial)
	C.fluidaudio_free_string(cPartial)
	return &text, nil
}

// Qwen3StreamingFinish completes the Qwen3 streaming session and returns the final text.
func (f *FluidAudio) Qwen3StreamingFinish() (string, error) {
	var cText *C.char
	rc := C.fluidaudio_qwen3_streaming_finish(f.ptr, &cText)
	if rc != 0 {
		return "", &FluidAudioError{Code: int32(rc), Message: "qwen3_streaming_finish failed"}
	}
	text := C.GoString(cText)
	C.fluidaudio_free_string(cText)
	return text, nil
}

// IsQwen3StreamingAvailable returns whether the Qwen3 streaming engine is initialized.
func (f *FluidAudio) IsQwen3StreamingAvailable() bool {
	return C.fluidaudio_is_qwen3_streaming_available(f.ptr) == 1
}

// --- System Info ---

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
