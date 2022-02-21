package nitrotype

// CalculateWPM returns the Avg WPM using the total seconds and typed characters.
// Nitro Type Team Page rounds this figure.
func CalculateWPM(typed int, secs int) float64 {
	if secs == 0 {
		return 0.0
	}
	return float64(typed) / 5.0 / (float64(secs) / 60.0)
}

// CalculateAccuracy returns the accuracy using typed characters and errors (used on Nitro Type Team page)
func CalculateAccuracy(typed int, errs int) float64 {
	if typed == 0 {
		return 0.0
	}
	return (1.0 - (float64(errs) / float64(typed))) * 100.0
}

// CalculatePoints returns the points earned using Nitro Type's formula
func CalculatePoints(races int, wpm float64, accuracy float64) float64 {
	if races == 0 {
		return 0.0
	}
	return ((100.0 + (wpm / 2)) * (accuracy / 100.0)) * float64(races)
}
