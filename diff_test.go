package hl7

import (
	"bytes"

	"github.com/mb0/diff"
)

func splitLine(b []byte) [][]byte {
	ff := bytes.FieldsFunc(b, func(r rune) bool {
		return r == '\n' || r == '\r'
	})
	return ff
}

func lineDiff(a, b []byte) []byte {
	l := lines{splitLine(a), splitLine(b)}
	cs := diff.Diff(len(l.a), len(l.b), diff.Data(l))

	var buf bytes.Buffer
	for _, c := range cs {
		if true && c.Ins == c.Del {
			ct := c.Ins
			for i := 0; i < ct; i++ {
				a := l.a[c.A+i]
				b := l.b[c.B+i]
				buf.WriteString("<> ")
				buf.Write(byteDiff(a, b))
				buf.WriteByte('\n')

				buf.WriteString(" - ")
				buf.Write(a)
				buf.WriteByte('\n')

				buf.WriteString(" + ")
				buf.Write(b)
				buf.WriteByte('\n')
				buf.WriteByte('\n')
			}
			continue
		}
		for _, del := range l.a[c.A : c.A+c.Del] {
			buf.WriteString(" - ")
			buf.Write(del)
			buf.WriteByte('\n')
		}
		for _, ins := range l.b[c.B : c.B+c.Ins] {
			buf.WriteString(" + ")
			buf.Write(ins)
			buf.WriteByte('\n')
		}
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}

type lines struct {
	a, b [][]byte
}

func (l lines) Equal(i, j int) bool {
	return bytes.Equal(l.a[i], l.b[j])
}

func byteDiff(a, b []byte) []byte {
	var buf bytes.Buffer
	n := 0
	for _, c := range diff.Granular(1, diff.Bytes(a, b)) {
		buf.Write(b[n:c.B])
		if c.Ins > 0 {
			buf.Write(green(b[c.B : c.B+c.Ins]))
		}
		if c.Del > 0 {
			buf.Write(red(a[c.A : c.A+c.Del]))
		}
		n = c.B + c.Ins
	}
	buf.Write(b[n:])
	return buf.Bytes()
}

func green(b []byte) []byte {
	var buf bytes.Buffer
	buf.WriteString("\x1b[1;92m")
	buf.Write(b)
	buf.WriteString("\x1b[0m")
	return buf.Bytes()
}

func red(b []byte) []byte {
	var buf bytes.Buffer
	buf.WriteString("\x1b[1;91m")
	buf.Write(b)
	buf.WriteString("\x1b[0m")
	return buf.Bytes()
}
