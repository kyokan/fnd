package config

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateDefaultConfigFile(t *testing.T) {
	generatedCfg := GenerateDefaultConfigFile()
	cfg, err := ReadConfig(bytes.NewReader(generatedCfg))
	require.NoError(t, err)
	require.EqualValues(t, DefaultConfig, *cfg)
}
