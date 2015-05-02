package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

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

const (
	// The size in bytes of the data contained in the main
	// metadata segment of a splice file
	numHeaderBytes         = 6
	numInitialPaddingBytes = 6
	numHWVersionBytes      = 32
	numTempoBytes          = 4

	// This value is at the beginning of every splice file
	headerValue = "SPLICE"
)

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
func readPattern(r io.Reader) (*Pattern, error) {
	// Read header
	header := make([]byte, numHeaderBytes)
	if _, err := r.Read(header); err != nil {
		return nil, mapDecodeError(err)
	}

	// Validate header
	if string(header) != headerValue {
		return nil, ErrInvalidSpliceData
	}

	// Read payload size
	var payloadSize uint64
	if err := binary.Read(r, binary.BigEndian, &payloadSize); err != nil {
		return nil, mapDecodeError(err)
	}

	// Read hardware version
	hwVersionBytes := make([]byte, numHWVersionBytes)
	if _, err := r.Read(hwVersionBytes); err != nil {
		return nil, mapDecodeError(err)
	}
	hwVersionBytes = removeNullBytes(hwVersionBytes)
	hwVersion := string(hwVersionBytes)

	// Read tempo
	var tempo float32
	if err := binary.Read(r, binary.LittleEndian, &tempo); err != nil {
		return nil, mapDecodeError(err)
	}

	// Read track data
	trackSize := trackDataSize(payloadSize)
	trackBytes := make([]byte, trackSize)
	if _, err := r.Read(trackBytes); err != nil {
		return nil, mapDecodeError(err)
	}

	// Parse bytes into tracks
	trackBuffer := bytes.NewBuffer(trackBytes)
	tracks, err := parseTracks(trackBuffer)
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
