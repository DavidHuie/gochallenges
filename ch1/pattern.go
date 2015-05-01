package drum

import "fmt"

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	HW     string
	Tempo  float32
	Tracks []*Track
}

func (p *Pattern) String() string {
	str := ""
	str += fmt.Sprintf("Saved with HW Version: %s\n", p.HW)
	str += fmt.Sprintf("Tempo: %v\n", p.Tempo)

	for _, track := range p.Tracks {
		str += fmt.Sprintf("(%v) %s\t", track.ID, track.Name)

		if len(track.Notes) > 0 {
			for i, note := range track.Notes {
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
	}
	return str
}
