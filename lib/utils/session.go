package utils

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

// If AccessType is role, then the Profile needs to be configured in your shared config file ~/.aws/config to use have the assume role config setup
// Example:
/*
[profile <profile name>]
role_arn = arn:aws:iam::123456789012:role/<role name>
source_profile = <source profile in ~/.aws/credentials>
region = us-east-1
output = json
*/
// If AccessType is profile, then it will just use the profile in your shared credential file ~/.aws/credentials
type AccountInfo struct {
	AccountId  string
	Arn        string // only required if AccessType is instanceassume
	ExternalId string
	AccessType string
	Profile    string
}

// GetAccountId will get the account ID for the profile currently in use for the session
func GetAccountId(sess *session.Session) (string, error) {
	params := &sts.GetCallerIdentityInput{}

	resp, err := sts.New(sess).GetCallerIdentity(params)
	if err != nil {
		return "", err
	}

	id := *resp.Account
	return id, nil
}

func (account AccountInfo) SetAccountId() error {
	sess, err := account.GetSession("us-east-1")
	if err != nil {
		return err
	}
	account.AccountId, err = GetAccountId(sess)
	if err != nil {
		return err
	}
	return nil
}

func (account AccountInfo) GetSession(region string) (*session.Session, error) {
	var sess *session.Session
	if region == "" {
		region = "us-east-1"
	}
	var err error
	switch account.AccessType {
	case "assume":
		sess = AssumeRoleWithProfile(account, region)
	case "profile":
		sess = OpenSession(account.Profile, region)
	case "instance":
		sess = session.Must(session.NewSession())
	case "instanceassume":
		sess, err = AssumeRoleWithInstance(account, region)
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

// Assumes the role of the specified profile
func AssumeRoleWithProfile(account AccountInfo, region string) *session.Session {
	LoadConfigFile()
	sess := session.Must(session.NewSessionWithOptions(session.Options{Config: aws.Config{Region: aws.String(region)}, Profile: account.Profile, SharedConfigState: session.SharedConfigEnable}))
	return sess
}

// Assumes the role of the given arn with the instance profile and returns a session into the account associated with the arn
func AssumeRoleWithInstance(account AccountInfo, region string) (*session.Session, error) {
	// open a new session with the instance profile
	sess := session.Must(session.NewSession())
	svc := sts.New(sess)

	params := &sts.AssumeRoleInput{
		RoleArn:         aws.String(account.Arn),
		RoleSessionName: aws.String(""),
		DurationSeconds: aws.Int64(900),
	}
	if account.ExternalId != "" {
		params.ExternalId = &account.ExternalId
	}

	// AssumeRole gets an Access Key, Secret Key, and Session Token into the client account with the provided arn
	resp, err := svc.AssumeRole(params)
	if err != nil {
		return nil, err
	}
	if resp.Credentials == nil {
		return nil, fmt.Errorf("could not assume role")
	}

	id := *resp.Credentials.AccessKeyId
	secret := *resp.Credentials.SecretAccessKey
	token := *resp.Credentials.SessionToken
	// NewStaticCredentials gives the new session the credentials to use when opening the new session, based on the credentials from the AssumeRole response
	newSess := session.Must(session.NewSessionWithOptions(session.Options{Config: aws.Config{Region: aws.String(region), Credentials: credentials.NewStaticCredentials(id, secret, token)}}))
	return newSess, nil
}

// This is a helper func to load the ~/.aws/config file
func LoadConfigFile() {
	os.Setenv("AWS_SDK_LOAD_CONFIG", strconv.Itoa(1))
}
