package disk

import (
	"github.com/stretchr/testify/require"
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
