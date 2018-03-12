package goaddtestcoverage

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

// Pipe takes the filename and an io Read and Writer. The filename is only used
// in naming the checkpoints. The input is parsed, checkpoints are added and the
// results are written to out.
func Pipe(filename string, in io.Reader, out io.Writer) error {
	op := &procOp{
		r:     bufio.NewReader(in),
		w:     out,
		token: make([]rune, 1, 20),
		file:  filename,
	}

	for op.err == nil {
		token := op.getToken()
		switch token {
		case `'`, `"`:
			op.stringState(op.char)
		case "/":
			op.slashState() // line comment, block comment or regex
		case "function", "if", "for", "while":
			op.functionalBlockState()
		}
	}
	if op.err == io.EOF {
		op.err = nil
	}

	if len(op.markers) > 0 {
		op.WriteString("Lapiz.Test.regMks(\n  ")
		op.WriteString(strings.Join(op.markers, ",\n  "))
		op.WriteString("\n);")
	}
	return op.err
}

type procOp struct {
	r            *bufio.Reader
	w            io.Writer
	err          error
	line, marker int
	char, peek   rune
	token        []rune
	file         string
	markers      []string
}

func (op *procOp) getPeek() {
	if op.err != nil || op.peek != 0 {
		return
	}
	op.peek, _, op.err = op.r.ReadRune()
	if op.peek == '\n' {
		op.line++
		op.marker = 0
	}
	if op.err == nil {
		op.WriteRune(op.peek)
	}
}

func (op *procOp) getChar() bool {
	if op.err != nil {
		return false
	}
	if op.peek == 0 {
		op.getPeek()
	}
	op.char, op.peek = op.peek, 0
	return op.err == nil
}

func (op *procOp) getToken() string {
	wasLetter := op.isLetter()
	op.getChar()
	op.token = op.token[:1]
	op.token[0] = op.char
	if wasLetter {
		for op.isLetter() && op.getChar() {
			op.token = append(op.token, op.char)
		}
	}
	return string(op.token)
}

func (op *procOp) isLetter() bool {
	op.getPeek()
	if op.err != nil {
		return false
	}
	return (op.peek >= 'a' && op.peek <= 'z') || (op.peek >= 'A' && op.peek <= 'Z')
}

// stringState is also used to check regex
// this will fail for /[/]/ but should probably use /[\/]/ anyway
func (op *procOp) stringState(quoteType rune) {
	skip := false
	for op.getChar() && (skip || op.char != quoteType) {
		if skip {
			skip = false
			continue
		}
		skip = op.char == '\\'
	}
}

func (op *procOp) slashState() {
	op.getChar()
	switch op.char {
	case '/':
		op.lineCommentState()
	case '*':
		op.blockCommentState()
	default:
		op.stringState('/')
	}
}

func (op *procOp) lineCommentState() {
	for op.getChar() && op.char != '\n' {
	}
}

func (op *procOp) blockCommentState() {
	endIfSlash := false
	for op.getChar() && !(endIfSlash && op.char == '/') {
		endIfSlash = op.char == '*'
	}
}

func (op *procOp) functionalBlockState() {
	parenDepth := 0
	for op.getChar() {
		switch op.char {
		case '(':
			parenDepth++
		case ')':
			parenDepth--
		case '{':
			if parenDepth == 0 {
				op.addMarker()
				return
			}
		}
	}
}

func (op *procOp) addMarker() {
	marker := strings.Join([]string{`"`, op.file, ") ", strconv.Itoa(op.line), " : ", strconv.Itoa(op.marker), `"`}, "")
	op.WriteString(`Lapiz.Test.incMk(`)
	op.WriteString(marker)
	op.WriteString(`);`)
	op.markers = append(op.markers, marker)
}

func (op *procOp) WriteString(str string) (int, error) {
	if op.err != nil {
		return 0, op.err
	}
	var n int
	n, op.err = op.w.Write([]byte(str))
	return n, op.err
}

func (op *procOp) WriteRune(r rune) (int, error) {
	return op.WriteString(string(r))
}
