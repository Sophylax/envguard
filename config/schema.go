package config

// Config configures envguard scanning behavior.
type Config struct {
	// EntropyThreshold is the minimum Shannon entropy required to report a token.
	EntropyThreshold float64 `json:"entropy_threshold" yaml:"entropy_threshold"`
	// MinLength is the minimum token length eligible for entropy analysis.
	MinLength int `json:"min_length" yaml:"min_length"`
	// MaxFileSizeKB is the maximum file size in kilobytes to scan before skipping with a warning.
	MaxFileSizeKB int `json:"max_file_size_kb" yaml:"max_file_size_kb"`
	// ExcludePaths contains glob patterns for files or directories that should not be scanned.
	ExcludePaths []string `json:"exclude_paths" yaml:"exclude_paths"`
	// ExcludeExtensions contains file extensions that should not be scanned.
	ExcludeExtensions []string `json:"exclude_extensions" yaml:"exclude_extensions"`
	// EntropyExcludePaths contains glob patterns for files or directories that should skip entropy scanning only.
	EntropyExcludePaths []string `json:"entropy_exclude_paths" yaml:"entropy_exclude_paths"`
	// CustomPatterns contains user-defined regex rules appended to the built-in pattern set.
	CustomPatterns []CustomPattern `json:"custom_patterns" yaml:"custom_patterns"`
}

// CustomPattern defines a user-provided regex-based secret detection rule.
type CustomPattern struct {
	// Name is the human-readable rule label.
	Name string `json:"name" yaml:"name"`
	// Pattern is the regular expression used for matching.
	Pattern string `json:"pattern" yaml:"pattern"`
	// Severity is the reporting level assigned to matches.
	Severity string `json:"severity" yaml:"severity"`
}

// Default returns the default envguard configuration.
func Default() Config {
	return Config{
		EntropyThreshold: 4.5,
		MinLength:        20,
		MaxFileSizeKB:    500,
		ExcludePaths: []string{
			"testdata/**",
			"**/*.test.js",
			"vendor/**",
		},
		ExcludeExtensions: []string{
			".lock",
			".svg",
			".png",
		},
	}
}
