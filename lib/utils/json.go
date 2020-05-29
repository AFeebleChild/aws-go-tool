package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
)

//Will take in a struct and print with pretty json
func PrettyPrintJson(input interface{}) {
	marshalled, _ := json.Marshal(input)
	var output bytes.Buffer
	json.Indent(&output, marshalled, "", "\t")
	fmt.Println(string(output.Bytes()))
}
