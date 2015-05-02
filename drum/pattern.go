package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
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
		str += fmt.Sprintf("(%v) %s\t", track.id, track.name)

		if len(track.notes) > 0 {
			for i, note := range track.notes {
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

const (
	// The size in bytes of the data contained in the main
	// metadata segment of a splice file
	numInitialPaddingBytes = 6
	numPayloadSizeBytes    = 8
	numHWVersionBytes      = 32
	numTempoBytes          = 4
)

// Returns the size of the track data payload based on
// the size of the entire payload.
func trackDataSize(payloadSize uint64) uint64 {
	return payloadSize -
		uint64(numHWVersionBytes) -
		uint64(numTempoBytes)
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

	if err := readPadding(file, numInitialPaddingBytes); err != nil {
		return nil, err
	}

	// Read payload size
	var payloadSize uint64
	if err := readIntoValue(
		file,
		numPayloadSizeBytes,
		&payloadSize,
		binary.BigEndian,
	); err != nil {
		return nil, err
	}

	// Read hardware version
	hwVersionBytes := make([]byte, numHWVersionBytes)
	if _, err := file.Read(hwVersionBytes); err != nil {
		return nil, err
	}
	hwVersionBytes = removeNullBytes(hwVersionBytes)
	hwVersion := string(hwVersionBytes)

	// Read tempo
	var tempo float32
	if err := readIntoValue(
		file,
		numTempoBytes,
		&tempo,
		binary.LittleEndian,
	); err != nil {
		return nil, err
	}

	trackSize := trackDataSize(payloadSize)

	// Read track data
	trackBytes := make([]byte, trackSize)
	if _, err := file.Read(trackBytes); err != nil {
		return nil, err
	}

	// Parse bytes into tracks
	trackBuffer := bytes.NewBuffer(trackBytes)
	tracks, err := parseTracks(trackBuffer)
	if err != nil {
		return nil, err
	}

	pattern := &Pattern{
		hw:     hwVersion,
		tempo:  tempo,
		tracks: tracks,
	}

	return pattern, nil
}
