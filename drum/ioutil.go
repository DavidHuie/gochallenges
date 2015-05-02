package drum

import "io"

// Reads n bytes from a reader, discarding the output.
func readPadding(r io.Reader, numBytes int) error {
	_, err := r.Read(make([]byte, numBytes))
	return err
}
