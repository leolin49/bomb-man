package common

import "math"

// εθδΊε₯
func Round(num float64) int {
	return int(math.Floor(num + 0.5))
}
