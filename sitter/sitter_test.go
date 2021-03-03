package sitter_test

import (
	"os"
	"strings"
	"testing"

	"github.com/carlmjohnson/exitcode"
	"github.com/matryer/is"
	"github.com/spotlightpa/kristy/sitter"
)

func TestCLI(t *testing.T) {
	cases := map[string]struct {
		in   string
		code int
	}{
		"help":     {"-h", 0},
		"longhelp": {"--help", 0},
		"nothing":  {"", 1},
		"hc":       {"-healthcheck - -slack -", 1},
		"ok":       {"-healthcheck $HEALTHCHECK -slack $SLACK echo hi", 0},
		"fail":     {"-healthcheck $HEALTHCHECK -slack $SLACK ls nonesuch", 1},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			is := is.New(t)
			in := os.ExpandEnv(tc.in)
			if len(in) < len(tc.in) {
				t.Skip("missing env var")
			}
			err := sitter.CLI(strings.Fields(in))
			is.Equal(exitcode.Get(err), tc.code) // exit code must match expectation
		})
	}
}
