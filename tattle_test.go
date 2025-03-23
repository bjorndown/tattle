package main

import (
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path"
	"testing"
)

func Test_parseDfOutput(t *testing.T) {
	dfOutput := `Mounted on                                                                                           Use%
/                                                                                                     69%
/dev                                                                                                   0%
/dev/shm                                                                                               1%
/boot                                                                            50%
/run                                                                                                   1%
`

	result := parseDfOutput(dfOutput)

	require.Equal(t, result, []Threshold{
		{
			Target:  "/",
			Percent: 69,
		}, {
			Target:  "/dev",
			Percent: 0,
		}, {
			Target:  "/dev/shm",
			Percent: 1,
		}, {
			Target:  "/boot",
			Percent: 50,
		}, {
			Target:  "/run",
			Percent: 1,
		},
	})
}

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
  "webhook": "https://some.com/webhook"
}`)

	require.Nil(t, err)

	config, err := readConfig(tempConfigFilePath)

	require.Nil(t, err)

	require.Equal(t, config, Config{
		Disk: DiskCheckConfig{
			Thresholds: []Threshold{
				{
					Target:  "/",
					Percent: 69,
				}, {
					Target:  "/home",
					Percent: 50,
				},
			},
		},
		WebhookUrl: "https://some.com/webhook",
	})
}
