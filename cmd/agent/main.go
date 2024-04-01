package main

import "github.com/smartfor/metrics/internal"

func main() {
	s := internal.NewService(nil)
	s.Run()
}
