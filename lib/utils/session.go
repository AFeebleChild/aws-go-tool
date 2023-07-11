package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
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
	Arn        string //only required if AccessType is instanceassume
	ExternalId string
	Region     string
	AccessType string
	Profile    string
}

//GetAccountId will get the account ID for the profile currently in use for the session
func GetAccountId(sess *session.Session) (string, error) {
	params := &sts.GetCallerIdentityInput{}

	resp, err := sts.New(sess).GetCallerIdentity(params)
	if err != nil {
		return "", err
	}

	id := *resp.Account
	return id, nil
}

func (account AccountInfo) GetSession() (*session.Session, error) {
	var sess *session.Session
	if account.Region == "" {
		account.Region = "us-east-1"
	}
	var err error
	switch account.AccessType {
	case "assume":
		sess = AssumeRoleWithProfile(account)
	case "profile":
		sess = OpenSession(account.Profile, account.Region)
	case "instance":
		sess = session.Must(session.NewSession())
	case "instanceassume":
		sess, err = AssumeRoleWithInstance(account)
	default:
		return nil, fmt.Errorf("no valid options in Access Type specified.  Needs 'assume', 'profile', 'instance', or 'instanceassume'")
	}
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func OpenSession(profile string, region string) *session.Session {
	sess := session.Must(session.NewSessionWithOptions(session.Options{Config: aws.Config{Region: aws.String(region)}, Profile: profile}))
	return sess
}

//Assumes the role of the specified profile
func AssumeRoleWithProfile(account AccountInfo) *session.Session {
	LoadConfigFile()
	sess := session.Must(session.NewSessionWithOptions(session.Options{Config: aws.Config{Region: aws.String(account.Region)}, Profile: account.Profile, SharedConfigState: session.SharedConfigEnable}))
	return sess
}

//Assumes the role of the given arn with the instance profile and returns a session into the account associated with the arn
func AssumeRoleWithInstance(account AccountInfo) (*session.Session, error) {
	//open a new session with the instance profile
	sess := session.Must(session.NewSession())
	svc := sts.New(sess)

	params := &sts.AssumeRoleInput{
		RoleArn:         aws.String(account.Arn), // Required
		RoleSessionName: aws.String(""),          // Required
		DurationSeconds: aws.Int64(900),
	}
	if account.ExternalId != "" {
		params.ExternalId = &account.ExternalId
	}

	//AssumeRole gets an Access Key, Secret Key, and Session Token into the client account with the provided arn
	resp, _ := svc.AssumeRole(params)

	if resp.Credentials == nil {
		log.Println("could not assume role properly")
		log.Println("Arn:", account.Arn)
		log.Println("Region:", account.Region)
		log.Println("SessionType:", account.AccessType)
		return nil, fmt.Errorf("could not assume role")
	}

	id := *resp.Credentials.AccessKeyId
	secret := *resp.Credentials.SecretAccessKey
	token := *resp.Credentials.SessionToken
	//NewStaticCredentials gives the new session the credentials to use when opening the new session, based on the credentials from the AssumeRole response
	newSess := session.Must(session.NewSessionWithOptions(session.Options{Config: aws.Config{Region: aws.String(account.Region), Credentials: credentials.NewStaticCredentials(id, secret, token)}}))
	return newSess, nil
}

//This is a helper func to load the ~/.aws/config file
func LoadConfigFile() {
	os.Setenv("AWS_SDK_LOAD_CONFIG", strconv.Itoa(1))
}
