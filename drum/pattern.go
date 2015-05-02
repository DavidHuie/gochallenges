package drum

import (
	"bytes"
	"fmt"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	hw     string
	tempo  float32
	tracks []*track
}

func (p *Pattern) String() string {
	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("Saved with HW Version: %s\n", p.hw))
	buf.WriteString(fmt.Sprintf("Tempo: %v\n", p.tempo))
	for _, track := range p.tracks {
		buf.WriteString(fmt.Sprintf("%v", track))
	}
	return buf.String()
}
