package utils

import (
	"math/rand"
	"strings"
	"time"
)

//Will generate a password with length of n using a custom random password generator
func GenPassword(n int) string {
	var x bool = true
	pass := make([]byte, n)

	//loop through the password generation until a valid password is generated
	for x {
		//letterBytes are the characters that will be used to generate the password
		const letterBytes = "!@#$%^&*()-_=+[]{}|0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		const (
			letterIdxBits = 6                    // 6 bits to represent a letter index
			letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
			letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
		)
		var src = rand.NewSource(time.Now().UnixNano())

		// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
		for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
			if remain == 0 {
				cache, remain = src.Int63(), letterIdxMax
			}
			if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
				pass[i] = letterBytes[idx]
				i--
			}
			cache >>= letterIdxBits
			remain--
		}

		//Checking to make sure the password will conform to the IAM password policy
		if strings.ContainsAny(string(pass), "!@#$%^&*()-_=+[]{}|0") &&
			strings.ContainsAny(string(pass), "0123456789") &&
			strings.ContainsAny(string(pass), "abcdefghijklmnopqrstuvwxyz") &&
			strings.ContainsAny(string(pass), "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			x = false
		} else {
			x = true
		}
	}

	return string(pass)
}
