package cmd

import (
	"os"
	"testing"
)

func TestIsInteractiveInputWithDevNull(t *testing.T) {
	file, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	defer file.Close()

	if isInteractiveInput(file) {
		t.Fatalf("%s should not be treated as interactive input", os.DevNull)
	}
}
