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

const (
	// The size in bytes of the data contained in the main
	// metadata segment of a splice file
	NumInitialPaddingBytes = 6
	NumPayloadSizeBytes    = 8
	NumHWVersionBytes      = 32
	NumTempoBytes          = 4
)

// Returns the size of the track data payload based on
// the size of the entire payload.
func TrackDataSize(payloadSize uint64) uint64 {
	return payloadSize -
		uint64(NumHWVersionBytes) -
		uint64(NumTempoBytes)
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

	if err := ReadPadding(file, NumInitialPaddingBytes); err != nil {
		return nil, err
	}

	// Read payload size
	var payloadSize uint64
	if err := ReadIntoValue(
		file,
		NumPayloadSizeBytes,
		&payloadSize,
		binary.BigEndian,
	); err != nil {
		return nil, err
	}

	// Read hardware version
	hwVersionBytes := make([]byte, NumHWVersionBytes)
	if _, err := file.Read(hwVersionBytes); err != nil {
		return nil, err
	}
	hwVersionBytes = RemoveNullBytes(hwVersionBytes)
	hwVersion := string(hwVersionBytes)

	// Read tempo
	var tempo float32
	if err := ReadIntoValue(
		file,
		NumTempoBytes,
		&tempo,
		binary.LittleEndian,
	); err != nil {
		return nil, err
	}

	trackSize := TrackDataSize(payloadSize)

	// Read track data
	trackBytes := make([]byte, trackSize)
	if _, err := file.Read(trackBytes); err != nil {
		return nil, err
	}

	// Parse bytes into tracks
	trackBuffer := bytes.NewBuffer(trackBytes)
	tracks, err := ParseTracks(trackBuffer)
	if err != nil {
		return nil, err
	}

	pattern := &Pattern{
		HW:     hwVersion,
		Tempo:  tempo,
		Tracks: tracks,
	}

	return pattern, nil
}
