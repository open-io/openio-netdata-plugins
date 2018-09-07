package util

import "testing"

func TestVolumeInfo(t *testing.T) {
	info, id, err := VolumeInfo("/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("info: %v, id: %v", info, id)
	// TODO verify output
}
