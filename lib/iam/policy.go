package iam

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

//two structs to help with printing the policy documents
type Document struct {
	Version   string      `json:"Version"`
	Statement []Statement `json:"Statement"`
}

type Statement struct {
	Action   []string `json:"Action"`
	Resource string   `json:"Resource"`
	Effect   string   `json:"Effect"`
	Sid      string   `json:"Sid"`
}

//The Policies and PolicyVersion need to have the same index to match up for later reference
type ProfilePolicies struct {
	Profile        string
	PolicyDetails  []iam.ManagedPolicyDetail
	PolicyVersions []iam.PolicyVersion
}

type ProfilesPolicies []ProfilePolicies

func GetProfilePolicies(sess *session.Session) (ProfilePolicies, error) {
	svc := iam.New(sess)
	var policies ProfilePolicies

	detailsParams := &iam.GetAccountAuthorizationDetailsInput{
		Filter: aws.StringSlice([]string{"LocalManagedPolicy"}),
	}

	x := true
	for x {
		resp, err := svc.GetAccountAuthorizationDetails(detailsParams)
		if err != nil {
			return ProfilePolicies{}, err
		}
		for _, policy := range resp.Policies {
			policies.PolicyDetails = append(policies.PolicyDetails, *policy)
		}

		//If the response is not truncated, exit loop. Otherwise, set the params marker to the response marker
		if !*resp.IsTruncated {
			x = false
		} else {
			detailsParams.Marker = resp.Marker
		}
	}

	for _, detail := range policies.PolicyDetails {
		versionParams := &iam.GetPolicyVersionInput{
			PolicyArn: detail.Arn,
			VersionId: detail.DefaultVersionId,
		}

		resp, err := svc.GetPolicyVersion(versionParams)
		if err != nil {
			log.Println("could not get policy version for", detail.PolicyName, ":", err)
		}
		policies.PolicyVersions = append(policies.PolicyVersions, *resp.PolicyVersion)
	}

	return policies, nil
}

func GetProfilesPolicies(accounts []utils.AccountInfo) (ProfilesPolicies, error) {
	profilesPoliciesChan := make(chan ProfilePolicies)
	var wg sync.WaitGroup

	for _, account := range accounts {
		wg.Add(1)
		go func(account utils.AccountInfo) {
			defer wg.Done()
			fmt.Println("Getting policies for profile:", account.Profile)
			sess, err := account.GetSession()
			if err != nil {
				log.Println("could not open session for ", account.Profile, " : ", err)
				return
			}
			profilePolicies, err := GetProfilePolicies(sess)
			if err != nil {
				log.Println("could not get profiles for ", account.Profile, " : ", err)
				return
			}
			profilePolicies.Profile = account.Profile

			profilesPoliciesChan <- profilePolicies
		}(account)
	}

	go func() {
		wg.Wait()
		close(profilesPoliciesChan)
	}()

	var profilesPolicies ProfilesPolicies
	for profilePolicies := range profilesPoliciesChan {
		profilesPolicies = append(profilesPolicies, profilePolicies)
	}

	return profilesPolicies, nil
}

func WriteProfilesPolicies(profilesPolicies ProfilesPolicies) error {
	outputDir := "output/iam/"
	utils.MakeDir(outputDir)
	outputFile := outputDir + "policies.csv"
	outfile, err := utils.CreateFile(outputFile)
	if err != nil {
		return fmt.Errorf("could not create policies file: %v", err)
	}
	writer := csv.NewWriter(outfile)
	defer writer.Flush()
	fmt.Println("Writing policies to file", outfile.Name())
	writer.Write([]string{"Account", "Policy", "Description", "Create Date", "Attachment Count"})

	for _, profilePolicies := range profilesPolicies {
		for x := 0; x < len(profilePolicies.PolicyVersions); x++ {
			policyName := *profilePolicies.PolicyDetails[x].PolicyName
			profile := profilePolicies.Profile

			//writing the policy document to file in a profile specific directory
			policyDir := outputDir + "/" + profilePolicies.Profile
			utils.MakeDir(policyDir)
			//filename to create for the policy document
			policyOutput := policyDir + "/" + policyName + ".json"
			file, err := os.Create(policyOutput)
			if err != nil {
				log.Println("could not open file for policy", policyName, "in account", profile, ":", err)
			}
			decoded, err := url.QueryUnescape(*profilePolicies.PolicyVersions[x].Document)
			if err != nil {
				log.Println("could not decode policy", policyName, "in account", profile, ":", err)
			}

			//decode the document and marshall into structs to be able to print
			document := &Document{}
			err = json.Unmarshal([]byte(decoded), document)
			if err != nil {
				log.Println("could not marshall policy", policyName, "in account", profile, ":", err)
			}
			enc := json.NewEncoder(file)
			enc.SetIndent("", "	")
			enc.Encode(document)

			//fmt.Fprintf(file, document)

			//write the policy details to the csv
			var description string
			if profilePolicies.PolicyDetails[x].Description != nil {
				description = *profilePolicies.PolicyDetails[x].Description
			}
			splitDate := strings.Split(profilePolicies.PolicyDetails[x].CreateDate.String(), " ")
			createDate := splitDate[0]
			var data = []string{
				profile,
				policyName,
				description,
				createDate,
				strconv.Itoa(int(*profilePolicies.PolicyDetails[x].AttachmentCount)),
			}

			err = writer.Write(data)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
