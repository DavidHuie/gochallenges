package drum

import (
	"bufio"
	"encoding/binary"
	"io"
)

const (
	// The size in bytes of the data contained in the
	// track segment of a splice file
	numTrackIDBytes       = 1
	numTrackPaddingBytes  = 3
	numTrackNameSizeBytes = 1

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

// Parses an entire io.Reader into a slice of Tracks.
func parseTracks(r io.Reader) ([]*track, error) {
	tracks := make([]*track, 0)
	b := bufio.NewReader(r)

	for {
		// Parse instrument number
		var trackID uint8
		err := readIntoValue(b, numTrackIDBytes,
			&trackID, binary.BigEndian)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if err := readPadding(b, numTrackPaddingBytes); err != nil {
			return nil, err
		}

		// Parse size of track name
		var trackNameSize uint8
		if err := readIntoValue(b, numTrackNameSizeBytes,
			&trackNameSize, binary.BigEndian); err != nil {
			return nil, err
		}

		// Parse track name
		trackNameBytes := make([]byte, trackNameSize)
		if _, err := b.Read(trackNameBytes); err != nil {
			return nil, err
		}
		trackName := string(trackNameBytes)

		// Parse notes
		noteBytes := make([]byte, numNotes)
		if _, err := b.Read(noteBytes); err != nil {
			return nil, err
		}
		notes := convertBytesToMeasure(noteBytes)

		track := &track{
			id:    trackID,
			name:  trackName,
			notes: notes,
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}
