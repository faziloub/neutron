package chunksplit_test

import (
	"bytes"
	"testing"

	"github.com/faziloub/neutron/backend/util/textproto/chunksplit"
)

func TestWriter(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{
			input:    []string{"helloworldhelloworldhelloworld"},
			expected: "hello/world/hello/world/hello/world",
		},
		{
			input:    []string{"helloworldhelloworldhe"},
			expected: "hello/world/hello/world/he",
		},
		{
			input:    []string{"helloworldhelloworldhe", "ll"},
			expected: "hello/world/hello/world/hell",
		},
		{
			input:    []string{"helloworldhelloworldhe", "llo"},
			expected: "hello/world/hello/world/hello",
		},
		{
			input:    []string{"helloworldhelloworldhe", "lloworld"},
			expected: "hello/world/hello/world/hello/world",
		},
		{
			input:    []string{"helloworldhelloworldhe", "llo", "wor", "ld"},
			expected: "hello/world/hello/world/hello/world",
		},
	}

	for _, test := range tests {
		var b bytes.Buffer
		w := chunksplit.New("/", 5, &b)

		for _, i := range test.input {
			w.Write([]byte(i))
		}

		output := b.String()
		if output != test.expected {
			t.Error("Got " + output + " instead of " + test.expected)
		}
	}
}
