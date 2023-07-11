package iam

import (
	"encoding/csv"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

type RoleInfo struct {
	Role             iam.Role
	InlinePolicies   []string
	AttachedPolicies []string
}

type ProfileRoles struct {
	Profile string
	Roles   []RoleInfo
}
type ProfilesRoles []ProfileRoles

func CreateRole(params *iam.CreateRoleInput, sess *session.Session) (*iam.CreateRoleOutput, error) {
	resp, err := iam.New(sess).CreateRole(params)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

//GetProfileRoles will get all the roles for a given profile session
func GetProfileRoles(sess *session.Session) ([]RoleInfo, error) {
	svc := iam.New(sess)
	params := &iam.ListRolesInput{
		MaxItems: aws.Int64(100),
	}

	var roles []iam.Role
	//x is the check to ensure there is no roles left from the IsTruncated
	x := true
	for x {
		resp, err := svc.ListRoles(params)
		if err != nil {
			return nil, fmt.Errorf("could not get roles: %v", err)
		}
		for _, role := range resp.Roles {
			roles = append(roles, *role)
		}
		//If it is truncated, add the marker to the params for the next loop
		//If not, set x to false to exit for loop
		if *resp.IsTruncated {
			params.Marker = resp.Marker
		} else {
			x = false
		}
	}

	var info []RoleInfo
	for _, role := range roles {
		inlineParams := &iam.ListRolePoliciesInput{
			RoleName: aws.String(*role.RoleName),
		}

		var tempInfo RoleInfo
		x := true
		for x {
			resp, err := svc.ListRolePolicies(inlineParams)
			if err != nil {
				return nil, fmt.Errorf("could not get inline role policies: %v", err)
			}

			tempInfo.Role = role
			for _, policy := range resp.PolicyNames {
				tempInfo.InlinePolicies = append(tempInfo.InlinePolicies, *policy)
			}
			if *resp.IsTruncated {
				params.Marker = resp.Marker
			} else {
				x = false
			}
		}

		attachedParams := &iam.ListAttachedRolePoliciesInput{
			RoleName: aws.String(*role.RoleName),
		}
		x = true
		for x {
			resp, err := svc.ListAttachedRolePolicies(attachedParams)
			if err != nil {
				return nil, fmt.Errorf("could not get inline role policies: %v", err)
			}

			tempInfo.Role = role
			for _, policy := range resp.AttachedPolicies {
				tempInfo.AttachedPolicies = append(tempInfo.AttachedPolicies, *policy.PolicyName)
			}
			if *resp.IsTruncated {
				params.Marker = resp.Marker
			} else {
				x = false
			}
		}
		info = append(info, tempInfo)
	}

	return info, nil
}

//GetProfilesRoles will get all of the roles in all given accounts
func GetProfilesRoles(accounts []utils.AccountInfo) (ProfilesRoles, error) {
	profilesRolesChan := make(chan ProfileRoles)
	var wg sync.WaitGroup

	for _, account := range accounts {
		wg.Add(1)
		go func(account utils.AccountInfo) {
			defer wg.Done()
			fmt.Println("Getting roles for profile:", account.Profile)
			var profileRoles ProfileRoles
			profileRoles.Profile = account.Profile
			sess, err := account.GetSession()
			if err != nil {
				log.Println("could not open session for ", account.Profile, " : ", err)
				return
			}
			roles, err := GetProfileRoles(sess)
			if err != nil {
				log.Println("could not get profiles for ", account.Profile, " : ", err)
				return
			}
			//for _, role := range roles {
			//	profileRoles.Roles = append(profileRoles.Roles, role)
			//}

			profileRoles.Roles = roles

			profilesRolesChan <- profileRoles
		}(account)
	}

	go func() {
		wg.Wait()
		close(profilesRolesChan)
	}()

	var profilesRoles ProfilesRoles
	for profileRoles := range profilesRolesChan {
		profilesRoles = append(profilesRoles, profileRoles)
	}

	return profilesRoles, nil
}

//func GetPolicyDocument(name string, sess *session.Session) {
//	params := &iam.getpolicy
//}

//UpdateProfilesRoles will take a filename which should be the output of the GetProfilesRoles func
//The duration parameter is the new MaxSessDuration in seconds
func UpdateProfilesRolesSessionDuration(filename string, duration int64) error {
	//TODO Update to use csv reader
	lines, err := utils.ReadFile(filename)
	if err != nil {
		return err
	}

	profileCompare := ""
	var sess *session.Session
	for x, line := range lines {
		splitLine := strings.Split(line, ",")
		profile, role := splitLine[0], splitLine[1]
		role = strings.Replace(role, " ", "", 1)
		//skip the first line as this should be the title of the columns
		if x == 0 {
			continue
		}
		//check if the last role is in the same account as the lastest, to reuse the session
		if profileCompare == "" {
			profileCompare = profile
			sess = utils.OpenSession(profile, "us-east-1")
		} else if profileCompare != "" && profile != profileCompare {
			profileCompare = profile
			sess = utils.OpenSession(profile, "us-east-1")
		}

		params := &iam.UpdateRoleInput{
			RoleName:           aws.String(role),
			MaxSessionDuration: aws.Int64(duration),
		}

		fmt.Printf("In profile %s, updating role %s\n", profile, role)
		_, err := iam.New(sess).UpdateRole(params)
		if err != nil {
			utils.LogAll("Could not update role", role, "in profile", profile, ":", err)
		}
	}
	return nil
}

func WriteProfilesRoles(profilesRoles ProfilesRoles) error {
	outputDir := "output/iam/"
	utils.MakeDir(outputDir)
	outputFile := outputDir + "roles.csv"
	outfile, err := utils.CreateFile(outputFile)
	if err != nil {
		return fmt.Errorf("could not create roles file: %v", err)
	}
	writer := csv.NewWriter(outfile)
	defer writer.Flush()
	fmt.Println("Writing roles to file", outfile.Name())
	writer.Write([]string{"Account", "Role", "Max Session Duration", "Attached Policies", "Inline Policies"})

	for _, profileRoles := range profilesRoles {
		for _, roleInfo := range profileRoles.Roles {
			stringAttached := strings.Join(roleInfo.AttachedPolicies, "|")
			stringInline := strings.Join(roleInfo.InlinePolicies, "|")

			var data = []string{
				profileRoles.Profile,
				*roleInfo.Role.RoleName,
				strconv.Itoa(int(*roleInfo.Role.MaxSessionDuration)),
				stringAttached,
				stringInline,
			}

			err = writer.Write(data)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
