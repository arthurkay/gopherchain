package utils

import (
	"fmt"
	"log"
	"runtime"
)

func HandleError(err error) {
	if err != nil {
		_, fn, line, _ := runtime.Caller(1)
		log.Printf("%s %s on line number %d %v", ANSIColor("\033[1;31m%s\033[0m", "[ERROR] "), fn, line, ANSIColor("\033[1;31m%s\033[0m", err.Error()))
	}
}

func ANSIColor(color, text string) string {
	return fmt.Sprintf(color, text)
}
