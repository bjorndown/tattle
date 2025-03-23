package disk

import (
	"fmt"
	"github.com/bjorndown/tattle/internal/common"
	"log/slog"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type CheckConfig struct {
	Thresholds []Threshold `json:"thresholds"`
}

type Threshold struct {
	Target  string `json:"target"`
	Percent int64  `json:"percent"`
}

var matchMoreThanTwoSpaces = regexp.MustCompile(" {2,}")

func CheckDiskSpace(config CheckConfig) (map[common.Status][]Threshold, error) {
	cmd := exec.Command("df", "--output=target,pcent")
	var out strings.Builder
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("running df failed: %w", err)
	}
	reportedThresholds := parseDfOutput(out.String())
	statusMap := make(map[common.Status][]Threshold)
	statusMap[common.OK] = make([]Threshold, 0)
	statusMap[common.NOK] = make([]Threshold, 0)
	for _, configuredThreshold := range config.Thresholds {
		for _, reportedThreshold := range reportedThresholds {
			if reportedThreshold.Target == configuredThreshold.Target && reportedThreshold.Percent+2 >= configuredThreshold.Percent {
				statusMap[common.NOK] = append(statusMap[common.NOK], configuredThreshold)
			} else {
				statusMap[common.OK] = append(statusMap[common.OK], configuredThreshold)
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

func GetMessageText(statusMap map[common.Status][]Threshold) []string {
	lines := []string{"## Disk space warnings\n"}

	for _, threshold := range statusMap[common.NOK] {
		lines = append(lines, fmt.Sprintf(" * mount point %q is close to threshold of %d%%", threshold.Target, threshold.Percent))
	}

	return lines
}
