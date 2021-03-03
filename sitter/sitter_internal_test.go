package sitter

import (
	"testing"

	"github.com/matryer/is"
)

func TestMakeMessage(t *testing.T) {
	cases := map[string]struct {
		in, err string
		limit   int
		msg     string
	}{
		"simple":       {"a", "b", 10, "-- stdout --\na\n-- stderr --\nb\n"},
		"no a":         {"", "123456", 5, "-- stdout --\n\n-- stderr --\n23456\n"},
		"lot of a":     {"12345", "1", 5, "-- stdout --\n2345\n-- stderr --\n1\n"},
		"more of a":    {"12345", "12", 5, "-- stdout --\n45\n-- stderr --\n12\n"},
		"no b":         {"12345", "", 5, "-- stdout --\n12345\n-- stderr --\n\n"},
		"lot of b":     {"1", "12345", 5, "-- stdout --\n1\n-- stderr --\n2345\n"},
		"more of b":    {"12", "12345", 5, "-- stdout --\n12\n-- stderr --\n45\n"},
		"lot of both":  {"12345", "12345", 10, "-- stdout --\n12345\n-- stderr --\n12345\n"},
		"more of both": {"12345678", "12345678", 10, "-- stdout --\n45678\n-- stderr --\n45678\n"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			is := is.New(t)
			limit := len("-- stdout --\n\n-- stderr --\n\n") + tc.limit
			msg := makeMessage([]byte(tc.in), []byte(tc.err), limit)
			is.Equal(tc.msg, string(msg))
		})
	}
}
