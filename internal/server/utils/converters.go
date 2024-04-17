package utils

import "strconv"

func GaugeAsString(v interface{}) string {
	return strconv.FormatFloat(v.(float64), 'f', -1, 64)
}

func CounterAsString(v interface{}) string {
	return strconv.FormatInt(v.(int64), 10)
}
