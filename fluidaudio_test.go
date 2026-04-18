package fluidaudio

import (
	"testing"
)

func TestNew(t *testing.T) {
	fa, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer fa.Close()

	if fa.ptr == nil {
		t.Fatal("bridge pointer is nil")
	}
}

func TestCloseIdempotent(t *testing.T) {
	fa, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	fa.Close()
	fa.Close() // should not panic
}

func TestSystemInfo(t *testing.T) {
	fa, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer fa.Close()

	info := fa.SystemInfo()

	if info.Platform == "" {
		t.Error("Platform is empty")
	}
	if info.ChipName == "" {
		t.Error("ChipName is empty")
	}
	if info.MemoryGB <= 0 {
		t.Errorf("MemoryGB should be positive, got %f", info.MemoryGB)
	}

	t.Logf("Platform: %s, Chip: %s, Memory: %.1f GB, AppleSilicon: %v",
		info.Platform, info.ChipName, info.MemoryGB, info.AppleSilicon)
}

func TestIsAppleSilicon(t *testing.T) {
	fa, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer fa.Close()

	result := fa.IsAppleSilicon()
	t.Logf("IsAppleSilicon: %v", result)
}
