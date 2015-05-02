package drum

type measure []note

// Converts a array of bytes consisting of the values
// 0x00 or 0x01 to measure.
func convertBytesToMeasure(b []byte) measure {
	var measure measure
	for _, n := range b {
		var val bool
		if n == 0x01 {
			val = true
		}
		measure = append(measure, note(val))
	}
	return measure
}
