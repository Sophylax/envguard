package scanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShannonEntropy(t *testing.T) {
	t.Run("high entropy aws like token", func(t *testing.T) {
		value := "AKIA7sD9LmQ2XpV8NcR4TyZ1WfK6HuB3Ja"
		assert.Greater(t, ShannonEntropy(value), 4.5)
	})

	t.Run("plain english text", func(t *testing.T) {
		value := "thisisaplainenglishsentence"
		assert.Less(t, ShannonEntropy(value), 3.5)
	})
}
