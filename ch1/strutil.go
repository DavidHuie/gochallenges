package drum

// Returns true if the input byte is a printable
// ASCII character.
func IsPrintableAscii(b byte) bool {
	return 32 <= b && b <= 126
}

// Removes all null bytes (0x00) from a byte array.
func RemoveNullBytes(b []byte) []byte {
	newB := make([]byte, 0)
	for _, byte := range b {
		if byte != 0 {
			newB = append(newB, byte)
		}
	}
	return newB
}
