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
	"slices"
	"strconv"
	"strings"
)

type Config struct {
	Disk       DiskCheckConfig    `json:"disk"`
	Systemd    SystemdCheckConfig `json:"systemd"`
	WebhookUrl string             `json:"webhook"`
}

type DiskCheckConfig struct {
	Thresholds []Threshold `json:"thresholds"`
}

type Threshold struct {
	Target  string `json:"target"`
	Percent int64  `json:"percent"`
}

type SystemdCheckConfig struct {
	ActiveUnits []string `json:"activeUnits"`
}

type SystemdUnitState struct {
	name  string
	state string
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

func checkSystemdUnits(config SystemdCheckConfig) (map[Status][]string, error) {
	statusMap := make(map[Status][]string)
	statusMap[OK] = make([]string, 0)
	statusMap[NOK] = make([]string, 0)

	for _, activeUnit := range config.ActiveUnits {
		cmd := exec.Command("systemctl", "--user", "--property=ActiveState", "show", activeUnit)
		var out strings.Builder
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("running systemctl failed: %w", err)
		}
		unitState, err := parseSystemCtlOutput(out.String(), activeUnit)
		if err != nil || unitState.state != "active" {
			statusMap[NOK] = append(statusMap[NOK], activeUnit)
		} else {
			statusMap[OK] = append(statusMap[OK], activeUnit)
		}
	}

	return statusMap, nil
}

func parseSystemCtlOutput(output string, unitName string) (SystemdUnitState, error) {
	tokens := strings.Split(output, "=")
	if len(tokens) != 2 {
		return SystemdUnitState{}, fmt.Errorf("cannot parse systemctl output: %q", tokens)
	}

	if tokens[0] != "ActiveState" {
		return SystemdUnitState{}, fmt.Errorf("got wrong property from systemctl: %q", tokens[0])
	}

	return SystemdUnitState{
		name:  unitName,
		state: tokens[1],
	}, nil
}

func getSystemdUnitsText(statusMap map[Status][]string) []string {
	lines := []string{"## Systemd unit warnings\n"}

	for _, unit := range statusMap[NOK] {
		lines = append(lines, fmt.Sprintf(" * systemd unit %q is not in active state", unit))
	}

	return lines
}

func checkDiskSpace(config DiskCheckConfig) (map[Status][]Threshold, error) {
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
	for _, configuredThreshold := range config.Thresholds {
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

func getDiskSpaceCheckText(statusMap map[Status][]Threshold) []string {
	lines := []string{"## Disk space warnings\n"}

	for _, threshold := range statusMap[NOK] {
		lines = append(lines, fmt.Sprintf(" * mount point %q is close to threshold of %d%%", threshold.Target, threshold.Percent))
	}

	return lines
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

func main() {
	flag.Parse()

	config, err := readConfig(configFilePath)

	if err != nil {
		panic(err)
	}

	statusMap, err := checkDiskSpace(config.Disk)

	if err != nil {
		panic(err)
	}

	lines := getDiskSpaceCheckText(statusMap)

	systemdStatusMap, err := checkSystemdUnits(config.Systemd)
	if err != nil {
		panic(err)
	}

	lines = slices.Concat(lines, getSystemdUnitsText(systemdStatusMap))

	if len(lines) > 0 {
		message := strings.Join(lines, "\n")
		err := sendMessage(message, config)
		if err != nil {
			panic(fmt.Sprintf("sending message failed: %v", err))
		}
		slog.Warn("check NOK", "check", "df")
	}
}
