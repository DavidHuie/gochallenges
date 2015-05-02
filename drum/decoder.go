package drum

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"os"
)

const (
	// The size in bytes of the data contained in a splice file
	numHeaderBytes         = 6
	numInitialPaddingBytes = 6
	numHWVersionBytes      = 32
	numTempoBytes          = 4
	numTrackPaddingBytes   = 3
	numNoteBytes           = 16

	// This value is at the beginning of every splice file
	headerValue = "SPLICE"
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
	pattern, err := d.readPattern()
	if err != nil {
		return err
	}
	*p = *pattern
	return nil
}

// Returns the size of the track data payload based on
// the size of the entire payload.
func trackDataSize(payloadSize uint64) uint64 {
	return payloadSize -
		uint64(numHWVersionBytes) -
		uint64(numTempoBytes)
}

// Maps a decoding error to an error that may be useful outside
// of the package.
func mapDecodeError(e error) error {
	if e == io.EOF {
		return ErrInvalidSpliceData
	}
	return e
}

// ErrInvalidSpliceData is returned when we detect an error
// in the binary structure of a splice file
var ErrInvalidSpliceData = errors.New("Splice file is invalid")

// Reads a pattern out of an io.Reader.
func (d *Decoder) readPattern() (*Pattern, error) {
	// Read header
	header := make([]byte, numHeaderBytes)
	if _, err := d.reader.Read(header); err != nil {
		return nil, mapDecodeError(err)
	}

	// Validate header
	if string(header) != headerValue {
		return nil, ErrInvalidSpliceData
	}

	// Read payload size
	var payloadSize uint64
	if err := binary.Read(d.reader, binary.BigEndian, &payloadSize); err != nil {
		return nil, mapDecodeError(err)
	}

	// Read hardware version
	hwVersionBytes := make([]byte, numHWVersionBytes)
	if _, err := d.reader.Read(hwVersionBytes); err != nil {
		return nil, mapDecodeError(err)
	}
	hwVersionBytes = removeNullBytes(hwVersionBytes)
	hwVersion := string(hwVersionBytes)

	// Read tempo
	var tempo float32
	if err := binary.Read(d.reader, binary.LittleEndian, &tempo); err != nil {
		return nil, mapDecodeError(err)
	}

	// Read track data
	tracks, err := d.parseTracks()
	if err != nil {
		return nil, mapDecodeError(err)
	}

	pattern := &Pattern{
		hw:     hwVersion,
		tempo:  tempo,
		tracks: tracks,
	}

	return pattern, nil
}

// Parses an entire io.Reader into a slice of Tracks.
func (d *Decoder) parseTracks() ([]*track, error) {
	var tracks []*track
	var parseError error

	for {
		// Parse instrument number
		var trackID uint8
		err := binary.Read(d.reader, binary.BigEndian, &trackID)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if err := readPadding(d.reader, numTrackPaddingBytes); err != nil {
			parseError = err
			break
		}

		// Parse size of track name
		var trackNameSize uint8
		if err := binary.Read(d.reader, binary.BigEndian, &trackNameSize); err != nil {
			parseError = err
			break
		}

		// Parse track name
		trackNameBytes := make([]byte, trackNameSize)
		if _, err := d.reader.Read(trackNameBytes); err != nil {
			parseError = err
			break
		}
		trackName := string(trackNameBytes)

		// Parse notes
		noteBytes := make([]byte, numNoteBytes)
		if _, err := d.reader.Read(noteBytes); err != nil {
			parseError = err
			break
		}
		notes := convertBytesToMeasure(noteBytes)

		track := &track{
			id:    trackID,
			name:  trackName,
			notes: notes,
		}
		tracks = append(tracks, track)
	}

	// If we were able to extract at least one track, don't return
	// an error.
	if len(tracks) == 0 && parseError != nil {
		log.Print("invalid track detected")
		return nil, parseError
	}

	return tracks, nil
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
