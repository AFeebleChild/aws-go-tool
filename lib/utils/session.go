package utils

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

//If AccessType is role, then the Profile needs to be configured in your shared config file ~/.aws/config to use have the assume role config setup
//Example:
/*
[profile <profile name>]
role_arn = arn:aws:iam::123456789012:role/<role name>
source_profile = <source profile in ~/.aws/credentials>
region = us-east-1
output = json
*/
//If AccessType is profile, then it will just use the profile in your shared credential file ~/.aws/credentials
type AccountInfo struct {
	AccountID  string
	Region     string
	AccessType string
	Profile    string
}

func OpenSession(profile string, region string) *session.Session {
	sess := session.Must(session.NewSessionWithOptions(session.Options{Config: aws.Config{Region: aws.String(region)}, Profile: profile}))
	return sess
}

//GetAccountID will get the account ID for the profile currently in use for the session
func GetAccountID(sess *session.Session) (string, error) {
	params := &sts.GetCallerIdentityInput{}

	resp, err := sts.New(sess).GetCallerIdentity(params)
	if err != nil {
		return "", err
	}

	id := *resp.Account
	return id, nil
}

func GetSession(account AccountInfo) (*session.Session, error) {
	var sess *session.Session
	if account.Region == "" {
		account.Region = "us-east-1"
	}
	switch account.AccessType {
	case "role":
		//fmt.Println("Assuming role for profile:", account.Profile)
		sess = AssumeClientRole(account)
	case "profile":
		//fmt.Println("Opening profile session for profile:", account.Profile)
		sess = OpenSession(account.Profile, account.Region)
	default:
		return nil, fmt.Errorf("no valid options in Access Type specified.  Needs 'role' or 'profile'")
	}
	return sess, nil
}

//Assumes the role give the specified profile
func AssumeClientRole(account AccountInfo) *session.Session {
	LoadConfigFile()
	sess := session.Must(session.NewSessionWithOptions(session.Options{Config: aws.Config{Region: aws.String(account.Region)}, Profile: account.Profile, SharedConfigState: session.SharedConfigEnable}))
	return sess
}

//This is a helper func to load the ~/.aws/config file
func LoadConfigFile() {
	os.Setenv("AWS_SDK_LOAD_CONFIG", string(1))
}
