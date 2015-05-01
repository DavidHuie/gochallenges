package drum

import (
	"bytes"
	"encoding/binary"
	"io"
)

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
