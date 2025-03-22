package main

import (
	"reflect"
	"testing"
)

func Test_parseDfOutput(t *testing.T) {
	type args struct {
		output string
	}
	tests := []struct {
		name string
		args args
		want []Threshold
	}{
		{
			name: "test",
			args: args{output: "Mounted on                                                                                           Use%\n/                                                                                                     69%\n/dev                                                                                                   0%\n/dev/shm                                                                                               1%\n/boot                                                                            50%\n/run                                                                                                   1%\n"},
			want: []Threshold{{
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseDfOutput(tt.args.output); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDfOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}
