package main

import "testing"

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
