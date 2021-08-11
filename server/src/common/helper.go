package common

import "math"

// 四舍五入
func Round(num float64) int {
	return int(math.Floor(num + 0.5))
}
