package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Config struct {
	Disk struct {
		Thresholds []Threshold `json:"thresholds"`
	} `json:"disk"`
	WebhookUrl string `json:"webhook"`
}

type Threshold struct {
	Target  string `json:"target"`
	Percent int64  `json:"percent"`
}

type Status uint8

const (
	OK Status = iota
	NOK
)

var matchMoreThanTwoSpaces = regexp.MustCompile(" {2,}")
var configFilePath string

func init() {
	flag.StringVar(&configFilePath, "config", "./config.json", "path to config file")
}

func CheckDiskSpace(config Config) (map[Status][]Threshold, error) {
	cmd := exec.Command("df", "--output=target,pcent")
	var out strings.Builder
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("running df failed: %w", err)
	}
	reportedThresholds := parseDfOutput(out.String())
	statusMap := make(map[Status][]Threshold)
	statusMap[OK] = make([]Threshold, 0)
	statusMap[NOK] = make([]Threshold, 0)
	for _, configuredThreshold := range config.Disk.Thresholds {
		for _, reportedThreshold := range reportedThresholds {
			if reportedThreshold.Target == configuredThreshold.Target && reportedThreshold.Percent+2 >= configuredThreshold.Percent {
				statusMap[NOK] = append(statusMap[NOK], configuredThreshold)
			} else {
				statusMap[OK] = append(statusMap[OK], configuredThreshold)
			}
		}
	}

	return statusMap, nil
}

func parseDfOutput(output string) []Threshold {
	var thresholds []Threshold
	lines := strings.Split(matchMoreThanTwoSpaces.ReplaceAllString(output, " "), "\n")
	for _, line := range lines {
		tokens := strings.Split(line, " ")
		if len(tokens) != 2 {
			continue
		}
		target := tokens[0]
		percentString := strings.Replace(tokens[1], "%", "", 1)
		percent, err := strconv.ParseInt(percentString, 10, 8)
		if err != nil {
			slog.Error("failed to parse to int", "error", err)
			continue
		}
		thresholds = append(thresholds, Threshold{
			Target:  target,
			Percent: percent,
		})
	}
	return thresholds
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
		slog.Error("couldnot ", "err", err)
		return Config{}, err
	}

	var config Config
	err = json.Unmarshal(file, &config)

	if err != nil {
		slog.Error("couldnot ", "err", err)
		return Config{}, err
	}

	return config, nil
}

func main() {
	flag.Parse()

	config, err := readConfig(configFilePath)

	if err != nil {
		panic(err)
	}

	statusMap, err := CheckDiskSpace(config)

	if err != nil {
		panic(err)
	}

	var lines = []string{"## Disk space warnings\n"}

	for _, threshold := range statusMap[NOK] {
		lines = append(lines, fmt.Sprintf(" * mount point %q is close to threshold of %d%%", threshold.Target, threshold.Percent))
	}

	if len(lines) > 0 {
		err := sendMessage(strings.Join(lines, "\n"), config)
		if err != nil {
			panic(fmt.Sprintf("sending message failed: %v", err))
		}
		slog.Warn("check NOK", "check", "df")
	}
}
