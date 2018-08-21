package vpc

import (
	"fmt"
	"os"

	"github.com/afeeblechild/aws-go-lib/utils"
)

var logFile *os.File

func init() {
	var err error
	logFile, err = utils.CreateFile("networkLog.txt")
	if err != nil {
		panic(err)
	}
	fmt.Println("Created log file:", logFile.Name())
}
