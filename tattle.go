package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bjorndown/tattle/internal/common"
	"github.com/bjorndown/tattle/internal/disk"
	"github.com/bjorndown/tattle/internal/systemd"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strings"
)

type Config struct {
	Disk       disk.CheckConfig    `json:"disk"`
	Systemd    systemd.CheckConfig `json:"systemd"`
	WebhookUrl string              `json:"webhook"`
}

var configFilePath string

func init() {
	flag.StringVar(&configFilePath, "config", "./config.json", "path to config file")
}

func sendMessage(text string, config Config) error {
	type Payload struct {
		Text string `json:"text"`
	}

	payload, err := json.Marshal(Payload{Text: text})
	if err != nil {
		return err
	}

	response, err := http.Post(config.WebhookUrl, "application/json", bytes.NewReader(payload))
	if err != nil {
		return err
	}

	if response.StatusCode != 201 {
		return fmt.Errorf("request failed with status=%d", response.StatusCode)
	}
	return nil
}

func readConfig(configFilePath string) (Config, error) {
	file, err := os.ReadFile(configFilePath)

	if err != nil {
		return Config{}, fmt.Errorf("cannot open config file: %w", err)
	}

	var config Config
	err = json.Unmarshal(file, &config)

	if err != nil {
		return Config{}, fmt.Errorf("cannot parse config file: %w", err)
	}

	return config, nil
}

func hasNok[I interface{}](statusMap map[common.Status][]I) bool {
	return len(statusMap[common.NOK]) > 0
}

func main() {
	flag.Parse()

	config, err := readConfig(configFilePath)

	if err != nil {
		panic(err)
	}

	diskStatusMap, err := disk.CheckDiskSpace(config.Disk)

	if hasNok(diskStatusMap) {
		slog.Warn("check NOK", "check", "disk")
	}

	if err != nil {
		panic(err)
	}

	lines := disk.GetMessageText(diskStatusMap)

	systemdStatusMap, err := systemd.CheckUnits(config.Systemd)
	if err != nil {
		panic(err)
	}

	if hasNok(systemdStatusMap) {
		slog.Warn("check NOK", "check", "systemd")
	}

	lines = slices.Concat(lines, systemd.GetMessageText(systemdStatusMap))

	if len(lines) > 0 {
		message := strings.Join(lines, "\n")
		err := sendMessage(message, config)
		if err != nil {
			panic(fmt.Sprintf("sending message failed: %v", err))
		}
		slog.Info("message sent")
	}
}
