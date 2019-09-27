package utils

import (
	"fmt"
	"log"
)

//Helper function to print to stdout and the log file
func LogAll(args ...interface{}) {
	fmt.Println(args...)
	log.Println(args...)
}
