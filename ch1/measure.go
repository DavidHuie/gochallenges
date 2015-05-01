package drum

type Measure []Note

// Converts a array of bytes consisting of the values
// 0x00 or 0x01 to an array of bools.
func ConvertBytesToMeasure(b []byte) Measure {
	measure := make(Measure, 0)
	for _, n := range b {
		var val bool
		if n == 0x01 {
			val = true
		}
		measure = append(measure, Note(val))
	}
	return measure
}
