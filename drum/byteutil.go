package drum

// Removes all null bytes (0x00) from a byte array.
func removeNullBytes(b []byte) []byte {
	newB := make([]byte, 0)
	for _, byte := range b {
		if byte != 0 {
			newB = append(newB, byte)
		}
	}
	return newB
}
