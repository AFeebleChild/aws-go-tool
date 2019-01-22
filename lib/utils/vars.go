package utils

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Sessioninfo struct {
	Sess    *session.Session
	Region  string
	Account string
}

type Ec2Options struct {
	Tags []string
}

var (
	RegionMap = []string{"us-east-1", "us-east-2", "us-west-1", "us-west-2", "ca-central-1", "eu-west-1", "eu-central-1", "eu-west-2", "ap-northeast-1", "ap-northeast-2", "ap-southeast-1", "ap-southeast-2", "ap-south-1", "sa-east-1"}
)

func BuildAccountsSlice(profilesFile string, accessType string) ([]AccountInfo, error) {
	profiles, err := ReadFile(profilesFile)
	if err != nil {
		return nil, fmt.Errorf("could not open profiles file", err)
	}

	var accounts []AccountInfo

	for _, profile := range profiles {
		var account AccountInfo
		account.Profile = profile
		account.AccessType = accessType
		accounts = append(accounts, account)
	}

	return accounts, nil
}
