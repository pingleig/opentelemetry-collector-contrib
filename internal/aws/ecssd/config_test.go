package ecssd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	b := mustReadFile(t, "testdata/config_example.yaml")
	c, err := LoadConfig(b)
	require.Nil(t, err)
	assert.Equal(t, ExampleConfig(), c)
}
