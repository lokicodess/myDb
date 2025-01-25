package utils

import "math/rand"

func RandInt() int {
	return rand.Intn(10000)
}

func Assert(b bool, message string) {
	if b {
		panic(message)
	}
}
