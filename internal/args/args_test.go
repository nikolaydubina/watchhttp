package args_test

import (
	"fmt"
	"testing"

	"golang.org/x/exp/slices"

	"github.com/nikolaydubina/watchhttp/internal/args"
)

func TestGetCommandFromArgs(t *testing.T) {
	tests := []struct {
		args     []string
		cmd      []string
		hasFlags bool
	}{
		{
			args:     nil,
			cmd:      nil,
			hasFlags: false,
		},
		{
			args:     []string{},
			cmd:      []string{},
			hasFlags: false,
		},
		{
			args:     []string{"--"},
			cmd:      []string{},
			hasFlags: false,
		},
		{
			args:     []string{"ls"},
			cmd:      []string{"ls"},
			hasFlags: false,
		},
		{
			args:     []string{"ls", "-la"},
			cmd:      []string{"ls", "-la"},
			hasFlags: false,
		},
		{
			args:     []string{"--", "ls"},
			cmd:      []string{"ls"},
			hasFlags: false,
		},
		{
			args:     []string{"-port", "9000", "--", "ls"},
			cmd:      []string{"ls"},
			hasFlags: true,
		},
		{
			args:     []string{"-port", "9000", "--", "ls", "-la"},
			cmd:      []string{"ls", "-la"},
			hasFlags: true,
		},
		{
			args:     []string{"-port", "9000", "--"},
			cmd:      nil,
			hasFlags: true,
		},
		{
			args:     []string{"-port", "9000"},
			cmd:      nil,
			hasFlags: true,
		},
		{
			args:     []string{"-h"},
			cmd:      nil,
			hasFlags: true,
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%v", tc.args), func(t *testing.T) {
			cmd, hasFlags := args.GetCommandFromArgs(tc.args)
			if !slices.Equal(tc.cmd, cmd) {
				t.Errorf("exp %v != %v", tc.cmd, cmd)
			}
			if tc.hasFlags != hasFlags {
				t.Errorf("exp %v != %v", tc.hasFlags, hasFlags)
			}
		})
	}
}
