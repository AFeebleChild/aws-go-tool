package ec2

import (
	"fmt"
	"encoding/csv"
	"strings"
	"sync"
	"log"
	"strconv"

	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws/session"
)

type RegionVolumes struct {
	Region    string
	Profile   string
	Volumes []ec2.Volume
}

type AccountVolumes []RegionVolumes
type ProfilesVolumes []AccountVolumes

//GetRegionVolumes will take a session and get all volumes based on the region of the session
func GetRegionVolumes(sess *session.Session) ([]ec2.Volume, error) {
	var volumes []ec2.Volume
	params := &ec2.DescribeVolumesInput{
		DryRun:   aws.Bool(false),
	}

	resp, err := ec2.New(sess).DescribeVolumes(params)
	if err != nil {
		return nil, err
	}

	//Add the volumes from the response to a slice to return
	for _, volume := range resp.Volumes {
		volumes = append(volumes, *volume)
	}

	return volumes, nil
}

//GetAccountVolumes will take a profile and go through all regions to get all volumes in the account
func GetAccountVolumes(account utils.AccountInfo) (AccountVolumes, error) {
	profile := account.Profile
	fmt.Println("Getting volumes for profile:", profile)
	volumesChan := make(chan RegionVolumes)
	var wg sync.WaitGroup

	for _, region := range utils.RegionMap {
		wg.Add(1)
		go func(region string) {
			var info RegionVolumes
			var err error
			defer wg.Done()
			account.Region = region
			sess, err := account.GetSession()
			if err != nil {
				log.Println("Could not get volumes for", account.Profile, ":", err)
				return
			}
			info.Volumes, err = GetRegionVolumes(sess)
			if err != nil {
				log.Println("Could not get volumes for", account.Profile, ":", err)
				return
			}
			info.Volumes, err = GetRegionVolumes(sess)
			if err != nil {
				log.Println("Could not get volumes for", account.Profile, ":", err)
			}
			info.Region = region
			info.Profile = profile
			volumesChan <- info
		}(region)
	}

	go func() {
		wg.Wait()
		close(volumesChan)
	}()

	var accountVolumes AccountVolumes
	for regionVolumes := range volumesChan {
		accountVolumes = append(accountVolumes, regionVolumes)
	}

	return accountVolumes, nil
}

//GetProfilesVolumes will return all the volumes in all accounts of a given filename with a list of profiles in it
func GetProfilesVolumes(accounts []utils.AccountInfo) (ProfilesVolumes, error) {
	profilesVolumesChan := make(chan AccountVolumes)
	var wg sync.WaitGroup

	//TODO need to add proper error handling for the go func
	for _, account := range accounts {
		wg.Add(1)
		go func(account utils.AccountInfo) {
			var err error
			defer wg.Done()
			accountVolumes, err := GetAccountVolumes(account)
			if err != nil {
				return
			}
			profilesVolumesChan <- accountVolumes
		}(account)
	}

	go func() {
		wg.Wait()
		close(profilesVolumesChan)
	}()

	var profilesVolumes ProfilesVolumes
	for accountVolumes := range profilesVolumesChan {
		profilesVolumes = append(profilesVolumes, accountVolumes)
	}
	return profilesVolumes, nil
}

func WriteProfilesVolumes(profileVolumes ProfilesVolumes, options utils.Ec2Options) error {
	outputDir := "output/ec2/"
	utils.MakeDir(outputDir)
	outputFile := outputDir + "volumes.csv"
	outfile, err := utils.CreateFile(outputFile)
	if err != nil {
		return fmt.Errorf("could not create volumes file", err)
	}
	writer := csv.NewWriter(outfile)
	defer writer.Flush()
	fmt.Println("Writing volumes to file:", outfile.Name())
	var columnTitles = []string{
		"Account",
		"Region",
		"Volume Name",
		"Volume ID",
		"Associated Instance",
		"Size",
		"State",
		"Create Date",
		"Encrypted",
		"KMS Key ID",
	}
	tags := options.Tags
	if len(tags) > 0 {
		for _, tag := range tags {
			columnTitles = append(columnTitles, tag)
		}
	}

	err = writer.Write(columnTitles)

	for _, accountVolumes := range profileVolumes {
		for _, regionVolumes := range accountVolumes {
			for _, volume := range regionVolumes.Volumes {
				var volumeName string
				for _, tag := range volume.Tags {
					if *tag.Key == "Name" {
						volumeName = *tag.Value
					}
				}
				//need to get kmsID separately as not all volumes will have one and trying to print *volume.KmsKeyId directly will fail if it doesn't have one
				var kmsID string
				if volume.KmsKeyId != nil {
					kmsID = *volume.KmsKeyId
				} else {
					kmsID = "nil"
				}

				volumeAttachment := "N/A"

				for _, attachment := range volume.Attachments {
					volumeAttachment = *attachment.InstanceId
				}

				//if *volume.VolumeId != "vol-ffffffff" {
				//	for _, volume := range regionVolumes.Volumes {
				//		if *volume.VolumeId == *volume.VolumeId {
				//			for _, attachment := range volume.Attachments{
				//				if *attachment.State == "attached" {
				//					volumeAttachment = *attachment.InstanceId
				//				}else {
				//					volumeAttachment = "unattached"
				//				}
				//			}
				//		}
				//	}
				//}

				splitDate := strings.Split(volume.CreateTime.String(), " ")
				createDate := splitDate[0]

				var data = []string{
					regionVolumes.Profile,
					regionVolumes.Region,
					volumeName,
					*volume.VolumeId,
					volumeAttachment,
					strconv.Itoa(int(*volume.Size)),
					*volume.State,
					createDate,
					strconv.FormatBool(*volume.Encrypted),
					kmsID,
				}

				if len(tags) > 0 {
					for _, tag := range tags {
						x := false
						for _, volumeTag := range volume.Tags {
							if *volumeTag.Key == tag {
								data = append(data, *volumeTag.Value)
								x = true
							}
						}
						if !x {
							data = append(data, "")
						}
					}
				}

				err = writer.Write(data)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	return nil
}