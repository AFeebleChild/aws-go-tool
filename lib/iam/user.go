package iam

import (
	"encoding/csv"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

type UserUpdate struct {
	Username      string
	ResetRequired bool
}

type ProfileUsers struct {
	Profile    string
	AccountID  string
	Users      []iam.User
	UsersInfo  []iam.UserDetail
	GroupsInfo []iam.GroupDetail
}
type ProfilesUsers []ProfileUsers

func UpdateUserPassword(user UserUpdate, sess *session.Session) (string, error) {
	svc := iam.New(sess)
	//The int passed to GenPassword is the length of the password generated
	password := utils.GenPassword(24)

	params := &iam.UpdateLoginProfileInput{
		UserName:              aws.String(user.Username),
		Password:              aws.String(password),
		PasswordResetRequired: aws.Bool(user.ResetRequired),
	}

	_, err := svc.UpdateLoginProfile(params)
	if err != nil {
		return "", err
	}
	return password, nil
}

//GetProfileUsers will get all the users for a given profile session
func GetProfileUsers(sess *session.Session) ([]iam.User, error) {
	svc := iam.New(sess)
	var users []iam.User
	params := &iam.ListUsersInput{
		MaxItems: aws.Int64(100),
	}

	//x is the check to ensure there is no users left from the IsTruncated
	x := true
	for x {
		resp, err := svc.ListUsers(params)
		if err != nil {
			return nil, err
		}
		for _, user := range resp.Users {
			users = append(users, *user)
		}
		//If it is truncated, add the marker to the params for the next loop
		//If not, set x to false to exit for loop
		if *resp.IsTruncated {
			params.Marker = resp.Marker
		} else {
			x = false
		}
	}

	return users, nil
}

func GetProfileAccountAuthInfo(sess *session.Session) ([]iam.GroupDetail, []iam.UserDetail, error) {
	svc := iam.New(sess)
	var groupinfo []iam.GroupDetail
	var userinfo []iam.UserDetail
	params := &iam.GetAccountAuthorizationDetailsInput{
		MaxItems: aws.Int64(100),
	}

	//x is the check to ensure there is no users left from the IsTruncated
	x := true
	for x {
		resp, err := svc.GetAccountAuthorizationDetails(params)
		if err != nil {
			return nil, nil, err
		}
		for _, group := range resp.GroupDetailList {
			groupinfo = append(groupinfo, *group)
		}
		for _, user := range resp.UserDetailList {
			userinfo = append(userinfo, *user)
		}
		//If it is truncated, add the marker to the params for the next loop
		//If not, set x to false to exit for loop
		if *resp.IsTruncated {
			params.Marker = resp.Marker
		} else {
			x = false
		}
	}

	return groupinfo, userinfo, nil
}

//GetProfilesUsers will get all of the users in all given accounts
func GetProfilesUsers(accounts []utils.AccountInfo) (ProfilesUsers, error) {
	profilesUsersChan := make(chan ProfileUsers)
	var wg sync.WaitGroup

	for _, account := range accounts {
		wg.Add(1)
		go func(account utils.AccountInfo) {
			fmt.Println("Getting users for profile:", account.Profile)
			defer wg.Done()
			var profileUsers ProfileUsers
			sess, err := account.GetSession()
			if err != nil {
				log.Println("Could not get users for", account.Profile, ":", err)
				return
			}
			users, err := GetProfileUsers(sess)
			if err != nil {
				log.Println("could not get users for ", account.Profile, " : ", err)
				return
			}
			for _, user := range users {
				profileUsers.Users = append(profileUsers.Users, user)
			}

			//Getting account auth info from iam to get policy info
			groupsinfo, usersinfo, err := GetProfileAccountAuthInfo(sess)
			for _, group := range groupsinfo {
				profileUsers.GroupsInfo = append(profileUsers.GroupsInfo, group)
			}
			for _, user := range usersinfo {
				profileUsers.UsersInfo = append(profileUsers.UsersInfo, user)
			}

			profileUsers.Profile = account.Profile
			profileUsers.AccountID, err = utils.GetAccountId(sess)
			if err != nil {
				log.Println("could not get account id for", account.Profile, ":", err)
				return
			}
			profilesUsersChan <- profileUsers
		}(account)
	}

	go func() {
		wg.Wait()
		close(profilesUsersChan)
	}()

	var profilesUsers ProfilesUsers
	for profileUsers := range profilesUsersChan {
		profilesUsers = append(profilesUsers, profileUsers)
	}

	return profilesUsers, nil
}

func WriteProfilesUsers(profilesUsers ProfilesUsers) error {
	outputDir := "output/iam/"
	utils.MakeDir(outputDir)
	outputFile := outputDir + "users.csv"
	outfile, err := utils.CreateFile(outputFile)
	if err != nil {
		return fmt.Errorf("could not create users file", err)
	}

	writer := csv.NewWriter(outfile)
	defer writer.Flush()
	fmt.Println("Writing users to file:", outfile.Name())
	var columnTitles = []string{"Account",
		"Account ID",
		"User Name",
		"Last Login Date",
		"Attached Policies",
		"Inline Policies",
		"Groups",
		"Group Policies",
		"Group Inline Policies",
	}

	err = writer.Write(columnTitles)

	for _, profileUsers := range profilesUsers {
		for _, user := range profileUsers.Users {
			var userInfo iam.UserDetail
			var policies, inlinePolicies []string
			for _, userinfo := range profileUsers.UsersInfo {
				if *userinfo.UserName == *user.UserName {
					userInfo = userinfo
					for _, policy := range userinfo.AttachedManagedPolicies {
						policies = append(policies, *policy.PolicyName)
					}
					for _, inlinePolicy := range userinfo.UserPolicyList {
						inlinePolicies = append(inlinePolicies, *inlinePolicy.PolicyName)
					}
				}
			}

			//Groups, Group Policies, Group Inline Policies
			var groups, groupPolicies, groupInlinePolicies []string
			for _, group := range userInfo.GroupList {
				groups = append(groups, *group)
			}
			for _, groupinfo := range profileUsers.GroupsInfo {
				for _, userGroup := range userInfo.GroupList {
					if *userGroup == *groupinfo.GroupName {
						for _, policy := range groupinfo.AttachedManagedPolicies {
							groupPolicies = append(groupPolicies, *policy.PolicyName)
						}
						for _, inlinepolicy := range groupinfo.GroupPolicyList {
							groupInlinePolicies = append(groupInlinePolicies, *inlinepolicy.PolicyName)
						}
					}
				}
			}
			var lastLogin string
			if user.PasswordLastUsed != nil {
				lastLogin = user.PasswordLastUsed.String()
				splitLogin := strings.Split(lastLogin, " ")
				lastLogin = splitLogin[0]
			} else {
				lastLogin = "N/A"
			}

			stringPolicies := strings.Join(policies, "|")
			stringInlinePolices := strings.Join(inlinePolicies, "|")
			stringGroups := strings.Join(groups, "|")
			stringGroupPolicies := strings.Join(groupPolicies, "|")
			stringGroupInlinePolicies := strings.Join(groupInlinePolicies, "|")

			var data = []string{profileUsers.Profile,
				profileUsers.AccountID,
				*user.UserName,
				lastLogin,
				stringPolicies,
				stringInlinePolices,
				stringGroups,
				stringGroupPolicies,
				stringGroupInlinePolicies,
			}

			err := writer.Write(data)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
