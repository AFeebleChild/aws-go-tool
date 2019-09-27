package utils

import (
	"bufio"
	"fmt"
	"log"
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
//It will check to ensure that it does not overwrite a file with the same name
func CreateFile(name string) (file *os.File, err error) {
	splitName := strings.Split(name, ".")
	lenSplit := len(splitName)
	//prefix is everything leading up to the last period in the name
	//suffix is everything after the last period in the name
	var prefix, suffix string
	if lenSplit == 1 {
		prefix = splitName[0]
	} else if lenSplit >= 2 {
		prefix, suffix = splitName[lenSplit-2], splitName[lenSplit-1]
	}

	x := true
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
		//Catch to break loop in case file reading goes wrong
		if i >= 1000 {
			fmt.Println("i >= 1000 in file name creation, breaking loop")
			x = false
		}
	}
	return
}

//This will create a new directory with the given path relative to where the script was run from
func MakeDir(path string) {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		if strings.Contains(err.Error(), "file exists") {
			return
		}
		//TODO add a return error
		log.Panic("could not create directory:", path, ":", err)
	}
}

//ReadFile will open a file, and return a string slice with each line as a string
func ReadFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		//Remove empty lines from the report
		if scanner.Text() == "" {
			continue
		}
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

//ParseELBLog will parse an elb log file and return the relevant details
//Will also check the Geo location of the IP if "geo" is true
func ParseELBLog(path string, geo bool) (ELBLogInfo, error) {
	var info ELBLogInfo
	lines, err := ReadFile(path)
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
			//TODO get top 5 IP locations in list, instead of all
			if geo {
				fmt.Println("Checking Geo Location for IP:", newIP)
				temp.SourceIPGeoInfo, _ = GetIPGeoLocation(newIP)
			}
			info = append(info, temp)
		}

	}

	return info, nil
}
