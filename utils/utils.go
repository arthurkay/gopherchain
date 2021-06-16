package utils

import (
	"log"
	"runtime"
)

func Logger(text string) {
	_, fn, line, _ := runtime.Caller(1)
	log.Printf("[Gopherchain] %s:%d %s", fn, line, text)
}

func Print(text string) {
	log.Printf("[Gopherchain] %s", text)
}
