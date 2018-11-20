package workspace

import (
	"encoding/csv"
	"fmt"
	"log"
	"sync"

	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"strconv"
)

type WorkspaceOptions struct {
	Tags []string
}

type RegionWorkspaces struct {
	AccountId    string
	Region       string
	Profile      string
	Instances    []workspaces.Workspace
	InstanceTags []workspaces.Tag
}

type AccountWorkspaces []RegionWorkspaces
type ProfilesWorkspaces []AccountWorkspaces

//GetRegionWorkspaces will take a session and pull all workspaces based on the region of the session
func GetRegionWorkspaces(sess *session.Session) ([]workspaces.Workspace, error) {
	params := &workspaces.DescribeWorkspacesInput{}

	var instances []workspaces.Workspace
	//x is the check to ensure there is no roles left from the IsTruncated
	x := true
	for x {
		resp, err := workspaces.New(sess).DescribeWorkspaces(params)
		if err != nil {
			return nil, fmt.Errorf("could not get workspaces", err)
		}
		for _, instance := range resp.Workspaces {
			instances = append(instances, *instance)
		}
		//If it is truncated, add the marker to the params for the next loop
		//If not, set x to false to exit for loop
		if resp.NextToken != nil {
			params.NextToken = resp.NextToken
		} else {
			x = false
		}
	}

	return instances, nil
}

func GetRegionWorkspaceTags(instances []workspaces.Workspace, sess *session.Session) ([]workspaces.Tag, error) {
	var tags []workspaces.Tag
	for instance := range instances {
		params := &workspaces.DescribeTagsInput{
			ResourceId: aws.String(*instance.WorkspaceId),
		}

		resp, err := workspaces.New(sess).DescribeTags(params)
	}

	return tags, nil
}

//GetAccountWorkspaces will take a profile and go through all regions to get all workspaces in the account
func GetAccountWorkspaces(account utils.AccountInfo) (AccountWorkspaces, error) {
	profile := account.Profile
	fmt.Println("Getting workspaces for profile:", profile)
	workspacesChan := make(chan RegionWorkspaces)
	var wg sync.WaitGroup

	for _, region := range utils.RegionMap {
		wg.Add(1)
		go func(region string) {
			var info RegionWorkspaces
			var err error
			defer wg.Done()
			account.Region = region
			sess, err := utils.GetSession(account)
			if err != nil {
				log.Println("could not get session for", account.Profile, ":", err)
				return
			}
			info.Instances, err = GetRegionWorkspaces(sess)
			if err != nil {
				log.Println("could not get workspaces for", region, "in", profile, ":", err)
				return
			}
			info.InstanceTags, err = GetRegionWorkspaceTags(info.Instances, sess)
			if err != nil {
				log.Println("could not get workspace tags for", region, "in", profile, ":", err)
				return
			}
			info.AccountId, err = utils.GetAccountID(sess)
			if err != nil {
				log.Println("could not get account id for profile: ", profile)
				return
			}
			info.Region = region
			info.Profile = profile
			workspacesChan <- info
		}(region)
	}

	go func() {
		wg.Wait()
		close(workspacesChan)
	}()

	var accountWorkspaces AccountWorkspaces
	for regionWorkspaces := range workspacesChan {
		accountWorkspaces = append(accountWorkspaces, regionWorkspaces)
	}

	return accountWorkspaces, nil
}

//GetProfilesWorkspaces will return all the workspaces in all accounts of a given filename with a list of profiles in it
func GetProfilesWorkspaces(accounts []utils.AccountInfo) (ProfilesWorkspaces, error) {
	profilesWorkspacesChan := make(chan AccountWorkspaces)
	var wg sync.WaitGroup

	for _, account := range accounts {
		wg.Add(1)
		go func(account utils.AccountInfo) {
			var err error
			defer wg.Done()
			accountWorkspaces, err := GetAccountWorkspaces(account)
			if err != nil {
				log.Println("Could not get workspaces for", account.Profile, ":", err)
				return
			}
			profilesWorkspacesChan <- accountWorkspaces
		}(account)
	}

	go func() {
		wg.Wait()
		close(profilesWorkspacesChan)
	}()

	var profilesWorkspaces ProfilesWorkspaces
	for accountWorkspaces := range profilesWorkspacesChan {
		profilesWorkspaces = append(profilesWorkspaces, accountWorkspaces)
	}
	return profilesWorkspaces, nil
}

func WriteProfilesWorkspaces(profileWorkspaces ProfilesWorkspaces, options utils.Ec2Options) error {
	outputDir := "output/workspaces/"
	utils.MakeDir(outputDir)
	outputFile := outputDir + "workspaces.csv"
	outfile, err := utils.CreateFile(outputFile)
	if err != nil {
		return fmt.Errorf("could not create workspaces file", err)
	}

	writer := csv.NewWriter(outfile)
	defer writer.Flush()
	fmt.Println("Writing workspaces to file:", outfile.Name())
	var columnTitles = []string{"Profile",
		"Account ID",
		"Region",
		"Workspace ID",
		"Computer Name",
		"Workspace State",
		"User Name",
		"Bundle ID",
		"Directory ID",
		"IP Address",
		"Root Volume Encrypted",
		"User Volume Encrypted",
		"Volume Encryption Key",
		"Subnet ID",
	}

	tags := options.Tags
	if len(tags) > 0 {
		for _, tag := range tags {
			columnTitles = append(columnTitles, tag)
		}
	}

	err = writer.Write(columnTitles)
	if err != nil {
		fmt.Println(err)
	}
	//var pemKeys []string
	//pemKeyFile, _ := utils.CreateFile("pemKeys.csv")
	for _, accountWorkspaces := range profileWorkspaces {
		for _, regionWorkspaces := range accountWorkspaces {
			for _, instance := range regionWorkspaces.Instances {
				var volumeEncryptionKey string
				if instance.VolumeEncryptionKey != nil {
					volumeEncryptionKey = *instance.VolumeEncryptionKey
				}
				var data = []string{regionWorkspaces.Profile,
					regionWorkspaces.AccountId,
					regionWorkspaces.Region,
					*instance.WorkspaceId,
					*instance.ComputerName,
					*instance.UserName,
					*instance.BundleId,
					*instance.DirectoryId,
					strconv.FormatBool(*instance.RootVolumeEncryptionEnabled),
					strconv.FormatBool(*instance.UserVolumeEncryptionEnabled),
					volumeEncryptionKey,
					*instance.SubnetId,
				}

				//if len(tags) > 0 {
				//	for _, tag := range tags {
				//		x := false
				//		for _, workspaceTag := range instance.WorkspaceProperties. {
				//			if *workspaceTag.Key == tag {
				//				data = append(data, *workspaceTag.Value)
				//				x = true
				//			}
				//		}
				//		if !x {
				//			data = append(data, "")
				//		}
				//	}
				//}

				err = writer.Write(data)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	return nil
}
