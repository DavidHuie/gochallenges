package drum

import (
	"bufio"
	"encoding/binary"
	"io"
)

const (
	// The size in bytes of the data contained in the
	// track segment of a splice file
	NumTrackIDBytes       = 1
	NumTrackPaddingBytes  = 3
	NumTrackNameSizeBytes = 1

	// The number of notes in each track
	NumNotes = 16
)

type Note bool

// A single instrument track for a drum machine.
type Track struct {
	ID    uint8
	Name  string
	Notes []Note
}

// Returns true if a slice looks like a track ID slice.
func IsID(b []byte) bool {
	return b[1] == 0 && b[2] == 0 && b[3] == 0
}

// Returns true if a slice looks like a track name slice.
func IsTrackNameSlice(b []byte) bool {
	return IsPrintableAscii(b[1]) &&
		IsPrintableAscii(b[2]) &&
		IsPrintableAscii(b[3])
}

// Parses an entire io.Reader into a slice of Tracks.
func ParseTracks(r io.Reader) ([]*Track, error) {
	tracks := make([]*Track, 0)
	b := bufio.NewReader(r)

	for {
		// Parse instrument number
		var trackID uint8
		err := ReadIntoValue(b, NumTrackIDBytes,
			&trackID, binary.BigEndian)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if err := ReadPadding(b, NumTrackPaddingBytes); err != nil {
			return nil, err
		}

		// Parse size of track name
		var trackNameSize uint8
		if err := ReadIntoValue(b, NumTrackNameSizeBytes,
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
		noteBytes := make([]byte, NumNotes)
		if _, err := b.Read(noteBytes); err != nil {
			return nil, err
		}
		notes := ConvertBytesToMeasure(noteBytes)

		track := &Track{
			ID:    trackID,
			Name:  trackName,
			Notes: notes,
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}
