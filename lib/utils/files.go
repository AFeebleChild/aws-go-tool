package utils

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type ELBIPInfo struct {
	SourceIP         string
	SourceIPCount    int
	SourceIPLocation string
	SourceIPGeoInfo  GeoIpInfo
}

type ELBLogInfo []ELBIPInfo

//CreateFile will create a file with a given name
//It will do checking to ensure that it does not overwrite a file with the same name
func CreateFile(name string) (file *os.File, err error) {
	x := true
	splitName := strings.Split(name, ".")
	var prefix, suffix string
	lenSplit := len(splitName)
	//TODO need to add logic if there is more than 1 period in the file name
	if lenSplit == 1 {
		prefix = splitName[0]
	} else if lenSplit == 2 {
		prefix, suffix = splitName[0], splitName[1]
	} else {
		err := errors.New("ERROR: unable to handle more than 1 period in file name")
		return nil, err
	}

	//Will loop through name appending to ensure that existing files with the same name are not overwritten
	for i := 0; x; i++ {
		if i == 0 {
			//this if will check to see if the file exists
			//if it does, continue the loop, otherwise create the file
			if _, err := os.Stat(name); err == nil {
				continue
			} else {
				file, err = os.Create(name)
				x = false
			}
		} else {
			if lenSplit == 1 {
				name = prefix + strconv.Itoa(i)
			} else {
				name = prefix + strconv.Itoa(i) + "." + suffix
			}
			if _, err := os.Stat(name); err == nil {
				continue
			} else {
				file, err = os.Create(name)
				x = false
			}
		}
		//Catch to break loop all in case file reading goes wrong
		if i >= 1000 {
			fmt.Println("i >= 1000, breaking loop")
			x = false
		}
	}
	return
}

//ReadFile will open a file, and return a string slice with each line as a string
//It is designed to be used for a file with a profile name on each line
func ReadProfilesFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

//ParseELBLog will parse an elb log file and return the relevant details
//Will also check the Geo location of the IP if "geo" is true
func ParseELBLog(path string, geo bool) (ELBLogInfo, error) {
	var info ELBLogInfo
	lines, err := ReadProfilesFile(path)
	if err != nil {
		return info, err
	}

	for _, line := range lines {
		//Split the log string and get the source IP address
		splitLine := strings.Split(line, " ")
		a := splitLine[2]
		//IPs in ELB logs are in the format <ip>:<port> so need to split again
		b := strings.Split(a, ":")
		newIP := b[0]

		//Search the current list of IPs and compare the IP of the new line
		//If it is an IP already encountered, add 1 to count, otherwise add to the list
		x := false
		for index, ipInfo := range info {
			if newIP == ipInfo.SourceIP {
				info[index].SourceIPCount += 1
				x = true
				continue
			}
		}
		if x == false {
			var temp ELBIPInfo
			temp.SourceIP = newIP
			temp.SourceIPCount = 1
			if geo {
				fmt.Println("Checking Geo Location for IP:", newIP)
				temp.SourceIPGeoInfo, _ = GetIPGeoLocation(newIP)
			}
			info = append(info, temp)
		}

	}

	return info, nil
}
