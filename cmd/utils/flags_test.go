package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCMDFlags_Init(t *testing.T) {
	assert := assert.New(t)
	flags := CMDFlags{}
	flags.Init()
	assert.Equal(flags.PDBMinAvailable, "2")
}
