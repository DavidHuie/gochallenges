package drum

import (
	"bytes"
	"fmt"
)

type note bool

func (n note) String() string {
	if n {
		return "x"
	}
	return "-"
}

// A single instrument track for a drum machine.
type track struct {
	id    uint8
	name  string
	notes []note
}

func (t *track) String() string {
	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("(%v) %s\t", t.id, t.name))

	if len(t.notes) > 0 {
		for i, note := range t.notes {
			if i%4 == 0 {
				buf.WriteString("|")
			}
			buf.WriteString(fmt.Sprintf("%v", note))
		}
		buf.WriteString("|\n")
	}

	return buf.String()
}
