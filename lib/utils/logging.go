package utils

import (
	"fmt"
	"log"
	"os"
)

//Helper function to print to stdout and the log file
func LogAll(args ...interface{}) {
	fmt.Println(args...)
	log.Println(args...)
}

func Check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}