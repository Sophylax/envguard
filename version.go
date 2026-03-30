package main

import "runtime/debug"

// resolveVersion prefers an injected build version, then falls back to module build info.
func resolveVersion(injected string) string {
	if injected != "" && injected != "dev" {
		return injected
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return injected
	}
	if info.Main.Version == "" || info.Main.Version == "(devel)" {
		return injected
	}
	return info.Main.Version
}
