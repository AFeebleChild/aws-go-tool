package utils

import (
	"io/ioutil"
	"net/http"
	"strings"
)

type GeoIpInfo struct {
	Country   string
	State     string
	City      string
	Latitude  string
	Longitude string
}

//GetIPGeoLocation will use api.hackertarget.com to find the location of an IP
func GetIPGeoLocation(ip string) (GeoIpInfo, error) {
	var info GeoIpInfo
	//crafting the api url in the form of
	//https://api.hackertarget.com/geoip/?q=1.1.1.1
	//TODO running into api limit on hackertarget.  need to see if http://ip-api.com is better
	//http://ip-api.com/#67.185.0.88
	url := "https://api.hackertarget.com/geoip/?q=" + ip
	resp, err := http.Get(url)
	if err != nil {
		return info, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	splitBody := strings.Split(string(body), "\n")

	for _, line := range splitBody {
		splitLine := strings.Split(line, ": ")
		var index, value string
		if len(splitLine) >= 2 {
			index = splitLine[0]
			value = splitLine[1]
		}
		if index == "Country" {
			info.Country = value
			continue
		}
		if index == "State" {
			info.State = value
			continue
		}
		if index == "City" {
			info.City = value
			continue
		}
		if index == "Latitude" {
			info.Latitude = value
			continue
		}
		if index == "Longitude" {
			info.Longitude = value
			continue
		}
	}

	return info, nil
}
