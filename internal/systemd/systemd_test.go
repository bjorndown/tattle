package systemd

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_parseSystemctlOutput(t *testing.T) {
	output, err := parseSystemCtlOutput("ActiveState=active", "foo.service")

	require.Nil(t, err)
	require.Equal(t, output, UnitState{
		name:  "foo.service",
		state: "active",
	})
}

func Test_parseSystemctlOutputError(t *testing.T) {
	_, err := parseSystemCtlOutput("OtherThing=foo", "foo.service")
	require.NotNil(t, err)
}
