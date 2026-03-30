package scanner

import "math"

// ShannonEntropy computes Shannon entropy in bits per character.
func ShannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	counts := make(map[rune]float64, len(s))
	for _, r := range s {
		counts[r]++
	}

	length := float64(len([]rune(s)))
	var entropy float64
	for _, count := range counts {
		p := count / length
		entropy -= p * math.Log2(p)
	}
	return entropy
}
