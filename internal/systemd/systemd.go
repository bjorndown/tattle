package systemd

import (
	"fmt"
	"github.com/bjorndown/tattle/internal/common"
	"os/exec"
	"strings"
)

type CheckConfig struct {
	ActiveUnits []string `json:"activeUnits"`
}

type UnitState struct {
	name  string
	state string
}

func CheckUnits(config CheckConfig) (map[common.Status][]string, error) {
	statusMap := make(map[common.Status][]string)
	statusMap[common.OK] = make([]string, 0)
	statusMap[common.NOK] = make([]string, 0)

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
			statusMap[common.NOK] = append(statusMap[common.NOK], activeUnit)
		} else {
			statusMap[common.OK] = append(statusMap[common.OK], activeUnit)
		}
	}

	return statusMap, nil
}

func parseSystemCtlOutput(output string, unitName string) (UnitState, error) {
	tokens := strings.Split(output, "=")
	if len(tokens) != 2 {
		return UnitState{}, fmt.Errorf("cannot parse systemctl output: %q", tokens)
	}

	if tokens[0] != "ActiveState" {
		return UnitState{}, fmt.Errorf("got wrong property from systemctl: %q", tokens[0])
	}

	return UnitState{
		name:  unitName,
		state: tokens[1],
	}, nil
}

func GetMessageText(statusMap map[common.Status][]string) []string {
	lines := []string{"## Systemd unit warnings\n"}

	for _, unit := range statusMap[common.NOK] {
		lines = append(lines, fmt.Sprintf(" * systemd unit %q is not in active state", unit))
	}

	return lines
}
