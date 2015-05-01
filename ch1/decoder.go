package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

// Binary file format
// 6 bytes: SPLICE in ASCII
// 7 bytes: padding
// 1 byte: C5 - 197 - number of remaining bytes
// 11 bytes: hw version
// 21 bytes: padding
// 4 bytes: 0000f042 -- float tempo
// 1 bytes: instrument number
// 3 bytes: padding
// 1 byte: 04 -- length in bytes of following instrument name
// 4 bytes: kick
// 16 bytes: notes

const (
	InitialPaddingSize  = 6
	PayloadSizeSize     = 8
	HWVersionSize       = 32
	TempoSize           = 4
	PreTrackPaddingSize = 4
	TrackNumSize        = 1
	TrackPaddingSize    = 3
	TrackNameSizeSize   = 1
)

// Returns the size of the track data payload based on
// the size of the entire payload.
func TrackDataSize(payloadSize int64) int64 {
	return payloadSize - int64(HWVersionSize) - int64(TempoSize)
}

// Reads n bytes from a reader, discarding the output.
func ReadPadding(r io.Reader, numBytes int) error {
	_, err := r.Read(make([]byte, numBytes))
	return err
}

// Read N bytes from a reader into a fixed-size value.
func ReadIntoValue(
	r io.Reader,
	numBytes int,
	value interface{},
	bo binary.ByteOrder) error {
	rawBytes := make([]byte, numBytes)
	if _, err := r.Read(rawBytes); err != nil {
		return err
	}
	buf := bytes.NewBuffer(rawBytes)

	if err := binary.Read(
		buf,
		bo,
		value,
	); err != nil {
		return err
	}

	return nil
}

// Validates whether a series of bytes corresponds
// to a measure of notes.
func IsMeasure(b []byte) bool {
	for _, n := range b {
		// TODO: this might be a weak assumption
		if n != 0 && n != 1 {
			return false
		}
	}
	return true
}

// Converts a array of bytes consisting of the values
// 0x00 or 0x01 to an array of bools.
func ConvertBytesToMeasure(b []byte) []bool {
	measure := make([]bool, 0)
	for _, n := range b {
		var val bool
		if n == 0x01 {
			val = true
		}
		measure = append(measure, val)
	}
	return measure
}

func IsID(b []byte) bool {
	return b[1] == 0 && b[2] == 0 && b[3] == 0
}

func IsPrintableAscii(b byte) bool {
	return 32 <= b && b <= 126
}

func IsTrackNameSlice(b []byte) bool {
	return IsPrintableAscii(b[1]) &&
		IsPrintableAscii(b[2]) &&
		IsPrintableAscii(b[3])
}

var ErrIsMetadata = errors.New("metadata detected")

func MeasureBytes(b *bufio.Reader) ([]byte, error) {
	measureBytes := make([]byte, 4)

	// We'll scan two measures to see if we're at the
	// metadata segment of the note data
	bytes, err := b.Peek(8)

	// If we get an EOF, we're certain we have a measure: the
	// last one
	if err == io.EOF {
		if _, err := b.Read(measureBytes); err != nil {
			return nil, err
		}
		return measureBytes, nil
	}

	measure := bytes[0:4]
	nextMeasure := bytes[4:8]

	if IsID(measure) && IsTrackNameSlice(nextMeasure) {
		return nil, ErrIsMetadata
	}

	if _, err = b.Read(measureBytes); err != nil {
		return nil, err
	}

	return measureBytes, nil
}

func RemoveNullBytes(b []byte) []byte {
	newB := make([]byte, 0)
	for _, byte := range b {
		if byte != 0 {
			newB = append(newB, byte)
		}
	}
	return newB
}

func ParseTracks(r io.Reader) ([]*Track, error) {
	tracks := make([]*Track, 0)
	bufr := bufio.NewReader(r)

	for {
		// Parse instrument number
		var trackNum uint8
		err := ReadIntoValue(bufr, TrackNumSize,
			&trackNum, binary.BigEndian)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if err := ReadPadding(bufr, TrackPaddingSize); err != nil {
			return nil, err
		}

		// Parse size of track name
		var trackNameSize int8
		if err := ReadIntoValue(bufr, TrackNameSizeSize,
			&trackNameSize, binary.BigEndian); err != nil {
			return nil, err
		}

		// Parse track name
		trackNameBytes := make([]byte, trackNameSize)
		if _, err := bufr.Read(trackNameBytes); err != nil {
			return nil, err
		}
		trackName := string(trackNameBytes)

		// Parse notes
		notes := make([]bool, 0)
		for {
			measureBytes, err := MeasureBytes(bufr)
			if err == ErrIsMetadata || err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}

			notes = append(notes,
				ConvertBytesToMeasure(measureBytes)...)
		}

		track := &Track{
			Num:   trackNum,
			Name:  trackName,
			Notes: notes,
		}
		tracks = append(tracks, track)
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

	if err := ReadPadding(file, InitialPaddingSize); err != nil {
		return nil, err
	}

	// Read payload size
	var payloadSize int64
	if err := ReadIntoValue(
		file,
		PayloadSizeSize,
		&payloadSize,
		binary.BigEndian,
	); err != nil {
		return nil, err
	}

	// Read hardware version
	hwVersionBytes := make([]byte, 32)
	if _, err := file.Read(hwVersionBytes); err != nil {
		return nil, err
	}
	hwVersionBytes = RemoveNullBytes(hwVersionBytes)
	hwVersion := string(hwVersionBytes)

	// Read tempo
	var tempo float32
	if err := ReadIntoValue(
		file,
		TempoSize,
		&tempo,
		binary.LittleEndian,
	); err != nil {
		return nil, err
	}

	// Calculate size of track data
	trackSize := TrackDataSize(payloadSize)

	// Read rest of track data
	trackBytes := make([]byte, trackSize)
	if _, err := file.Read(trackBytes); err != nil {
		return nil, err
	}
	trackBuffer := bytes.NewBuffer(trackBytes)

	// Parse tracks
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
		str += fmt.Sprintf("(%v) %s\t", track.Num, track.Name)

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

type Track struct {
	Num   uint8
	Name  string
	Notes []bool
}
