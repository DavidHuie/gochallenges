package drum

import (
	"io"
	"os"
)

// Decoder decodes a pattern from an io.Reader.
type Decoder struct {
	reader io.Reader
}

// NewDecoder returns a new Pattern decoder.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r}
}

// Decode decodes a pattern.
func (d *Decoder) Decode(p *Pattern) error {
	pattern, err := readPattern(d.reader)
	if err != nil {
		return err
	}
	*p = *pattern
	return nil
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var pattern Pattern
	decoder := NewDecoder(file)
	if err := decoder.Decode(&pattern); err != nil {
		return nil, err
	}

	return &pattern, nil
}
