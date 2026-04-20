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

	segments, err := fa.DiarizeOffline(testAudioFile)
	if err != nil {
		t.Fatalf("DiarizeFile() failed: %v", err)
	}

	t.Logf("Found %d segments", len(segments))
	for _, seg := range segments {
		t.Logf("  %d: %.2f-%.2f", seg.SpeakerID, seg.StartTime, seg.EndTime)
	}
}
