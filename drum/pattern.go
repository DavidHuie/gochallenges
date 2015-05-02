package drum

import "fmt"

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	hw     string
	tempo  float32
	tracks []*track
}

func (p *Pattern) String() string {
	str := ""
	str += fmt.Sprintf("Saved with HW Version: %s\n", p.hw)
	str += fmt.Sprintf("Tempo: %v\n", p.tempo)
	for _, track := range p.tracks {
		str += fmt.Sprintf("%v", track)
	}
	return str
}
