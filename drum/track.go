package drum

import "fmt"

const (
	// The size in bytes of the data contained in the
	// track segment of a splice file
	numTrackPaddingBytes = 3

	// The number of notes in each track
	numNotes = 16
)

type note bool

// A single instrument track for a drum machine.
type track struct {
	id    uint8
	name  string
	notes []note
}

func (t *track) String() string {
	str := fmt.Sprintf("(%v) %s\t", t.id, t.name)

	if len(t.notes) > 0 {
		for i, note := range t.notes {
			if i%4 == 0 {
				str += "|"
			}
			if note {
				str += "x"
			} else {
				str += "-"
			}
		}
		str += "|\n"
	}

	return str
}
