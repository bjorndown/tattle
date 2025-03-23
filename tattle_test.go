package main

import (
	"github.com/bjorndown/tattle/internal/disk"
	"github.com/bjorndown/tattle/internal/systemd"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path"
	"testing"
)

func createConfigFile(t *testing.T, content string) (string, error) {
	tempDir := t.TempDir()
	tempConfigFilePath := path.Join(tempDir, "config.json")
	configFile, err := os.Create(tempConfigFilePath)
	if err != nil {
		return "", err
	}
	_, err = io.WriteString(configFile, content)
	return tempConfigFilePath, nil
}

func Test_readConfig(t *testing.T) {
	tempConfigFilePath, err := createConfigFile(t, `{
  "disk": {
    "thresholds": [
      {
        "target": "/",
        "percent": 69
      },
      {
        "target": "/home",
        "percent": 50
      }
    ]
  },
  "systemd": {
    "activeUnits": ["foo.service"]
  },
  "webhook": "https://some.com/webhook"
}`)

	require.Nil(t, err)

	config, err := readConfig(tempConfigFilePath)

	require.Nil(t, err)

	require.Equal(t, config, Config{
		Disk: disk.CheckConfig{
			Thresholds: []disk.Threshold{
				{
					Target:  "/",
					Percent: 69,
				}, {
					Target:  "/home",
					Percent: 50,
				},
			},
		},
		Systemd:    systemd.CheckConfig{ActiveUnits: []string{"foo.service"}},
		WebhookUrl: "https://some.com/webhook",
	})
}
