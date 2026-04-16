//go:build integration

package fluidaudio

import (
	"os"
	"testing"
)

const testAudioFile = "testdata/test.wav"

func requireTestAudio(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(testAudioFile); os.IsNotExist(err) {
		t.Skipf("test audio file %s not found", testAudioFile)
	}
}

func TestTranscribeFile(t *testing.T) {
	requireTestAudio(t)

	fa, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer fa.Close()

	if err := fa.InitASR(); err != nil {
		t.Fatalf("InitASR() failed: %v", err)
	}

	result, err := fa.TranscribeFile(testAudioFile)
	if err != nil {
		t.Fatalf("TranscribeFile() failed: %v", err)
	}

	if result.Text == "" {
		t.Error("transcription text is empty")
	}
	t.Logf("Text: %s (confidence: %.2f, RTFx: %.2f)", result.Text, result.Confidence, result.RTFx)
}

func TestStreamingASR(t *testing.T) {
	requireTestAudio(t)

	fa, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer fa.Close()

	if err := fa.InitStreamingASR(); err != nil {
		t.Fatalf("InitStreamingASR() failed: %v", err)
	}

	result, err := fa.TranscribeFileStreaming(testAudioFile)
	if err != nil {
		t.Fatalf("TranscribeFileStreaming() failed: %v", err)
	}

	if result.Text == "" {
		t.Error("streaming transcription text is empty")
	}
	t.Logf("Text: %s (confidence: %.2f, RTFx: %.2f)", result.Text, result.Confidence, result.RTFx)
}

func TestDiarization(t *testing.T) {
	requireTestAudio(t)

	fa, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer fa.Close()

	if err := fa.InitDiarization(nil); err != nil {
		t.Fatalf("InitDiarization() failed: %v", err)
	}

	segments, err := fa.DiarizeFile(testAudioFile)
	if err != nil {
		t.Fatalf("DiarizeFile() failed: %v", err)
	}

	t.Logf("Found %d segments", len(segments))
	for _, seg := range segments {
		t.Logf("  %s: %.2f-%.2f (quality: %.2f)", seg.SpeakerID, seg.StartTime, seg.EndTime, seg.QualityScore)
	}
}

func TestQwen3TranscribeFile(t *testing.T) {
	requireTestAudio(t)

	fa, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer fa.Close()

	if err := fa.InitQwen3ASR(); err != nil {
		t.Skipf("InitQwen3ASR() failed (may require macOS 15+): %v", err)
	}

	result, err := fa.Qwen3TranscribeFile(testAudioFile, nil)
	if err != nil {
		t.Fatalf("Qwen3TranscribeFile() failed: %v", err)
	}

	if result.Text == "" {
		t.Error("Qwen3 transcription text is empty")
	}
	t.Logf("Text: %s (confidence: %.2f, RTFx: %.2f)", result.Text, result.Confidence, result.RTFx)
}
