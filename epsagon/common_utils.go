package epsagon

import "time"

func GetTimestamp() float64 {
	return float64(time.Now().UnixNano()) / float64(time.Millisecond) / float64(time.Nanosecond) / 1000.0
}
