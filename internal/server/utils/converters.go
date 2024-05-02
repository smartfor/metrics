package utils

import (
	"strconv"
)

func GaugeAsString(v interface{}) string {
	return strconv.FormatFloat(v.(float64), 'f', -1, 64)
}

func CounterAsString(v interface{}) string {
	return strconv.FormatInt(v.(int64), 10)
}

func GaugeFromString(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

func CounterFromString(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}
