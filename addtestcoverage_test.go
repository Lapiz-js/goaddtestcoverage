package goaddtestcoverage

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newTestProcOp(str string) (*procOp, *bytes.Buffer) {
	in := bytes.NewBufferString(str)
	out := &bytes.Buffer{}
	return &procOp{
		r:     bufio.NewReader(in),
		w:     out,
		token: make([]rune, 1, 20),
	}, out
}

func TestBlockCommentState(t *testing.T) {
	tt := map[string]string{
		"this is a test*/ this is not":         "this is a test*/",
		"this\nis\n***\na\ntest*/ this is not": "this\nis\n***\na\ntest*/",
	}
	for str, expected := range tt {
		t.Run(str, func(t *testing.T) {
			op, out := newTestProcOp(str)
			op.blockCommentState()
			assert.NoError(t, op.err)
			assert.Equal(t, expected, out.String())
		})
	}
}

func TestLineCommentState(t *testing.T) {
	tt := map[string]string{
		"this is a test\n this is not": "this is a test\n",
		"this*///foo\ntest":            "this*///foo\n",
	}
	for str, expected := range tt {
		t.Run(str, func(t *testing.T) {
			op, out := newTestProcOp(str)
			op.lineCommentState()
			assert.NoError(t, op.err)
			assert.Equal(t, expected, out.String())
		})
	}
}

func TestStringState(t *testing.T) {
	tt := map[string]string{
		`"this is a test"`: `"this is a test"`,
		`'this is a test'`: `'this is a test'`,
		`/this is a test/`: `/this is a test/`,
	}
	for str, expected := range tt {
		t.Run(str, func(t *testing.T) {
			op, out := newTestProcOp(str)
			op.getChar()
			op.stringState(op.char)
			assert.NoError(t, op.err)
			assert.Equal(t, expected, out.String())
		})
	}
}

func TestGetToken(t *testing.T) {
	op, _ := newTestProcOp(`Lapiz.Module(Filter, function($L)`)
	strs := []string{`Lapiz`, `.`, `Module`, `(`, `Filter`, `,`, ` `, `function`, `(`, `$`, `L`, `)`}
	for _, expected := range strs {
		if !assert.Equal(t, expected, op.getToken()) {
			break
		}
	}
}
