package main

import "os"

func trueCondition() bool {
	return true
}

func main() {
	if trueCondition() {
		os.Exit(1) // want "не используйте вызов os.Exit напрямую в main функции пакета main"
	}

	os.Exit(1) // want "не используйте вызов os.Exit напрямую в main функции пакета main"
}
