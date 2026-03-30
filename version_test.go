package main

import (
	"runtime/debug"
	"testing"
)

func TestResolveVersionPrefersInjectedVersion(t *testing.T) {
	t.Parallel()

	if got := resolveVersion("v1.2.3"); got != "v1.2.3" {
		t.Fatalf("resolveVersion() = %q, want %q", got, "v1.2.3")
	}
}

func TestResolveVersionFallsBackToInjectedDevWhenNoBuildInfoVersion(t *testing.T) {
	t.Parallel()

	if got := resolveVersion("dev"); got != "dev" {
		t.Fatalf("resolveVersion() = %q, want %q", got, "dev")
	}
}

func TestResolveVersionFallsBackToBuildInfoVersion(t *testing.T) {
	original := readBuildInfo
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Main: debug.Module{
				Version: "v0.1.1",
			},
		}, true
	}
	t.Cleanup(func() {
		readBuildInfo = original
	})

	if got := resolveVersion("dev"); got != "v0.1.1" {
		t.Fatalf("resolveVersion() = %q, want %q", got, "v0.1.1")
	}
}
