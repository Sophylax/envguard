package main

import "runtime/debug"

var readBuildInfo = debug.ReadBuildInfo

// resolveVersion prefers an injected build version, then falls back to module build info.
func resolveVersion(injected string) string {
	if injected != "" && injected != "dev" {
		return injected
	}

	info, ok := readBuildInfo()
	if !ok {
		return injected
	}
	if info.Main.Version == "" || info.Main.Version == "(devel)" {
		return injected
	}
	return info.Main.Version
}
