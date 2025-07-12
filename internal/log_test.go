package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validateTimeLayout(t *testing.T) {
	assert := assert.New(t)

	assert.NoError(ValidateTimeLayout("2006-01-02 15:04:05"))
	assert.Error(ValidateTimeLayout(""))
	assert.Error(ValidateTimeLayout("2006-01-02 15:04:0563:22"))
}
