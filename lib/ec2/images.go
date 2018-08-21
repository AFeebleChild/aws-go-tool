package ec2

import (
	"fmt"
	"sync"
	"log"
	"encoding/csv"

	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"strings"
)

type AmiOptions struct {
	Tags []string
}

type RegionImages struct {
	Profile   string
	AccountID string
	Region string
	Images      []ec2.Image
}

type AccountImages []RegionImages
type ProfilesImages []AccountImages

//GetRegionImages will take a session and pull all amis based on the region of the session
func GetRegionImages(sess *session.Session) ([]ec2.Image, error) {
	var amis []ec2.Image
	AccountID, err := utils.GetAccountID(sess)
	if err != nil {
		return amis, fmt.Errorf("could not get account id:", err)
	}
	params := &ec2.DescribeImagesInput{
		DryRun: aws.Bool(false),
		Owners: []*string{aws.String(AccountID)},
	}

	resp, err := ec2.New(sess).DescribeImages(params)
	if err != nil {
		return nil, err
	}

	for _, image := range resp.Images {
		amis = append(amis, *image)
	}

	return amis, nil
}

//GetAccountImages will take a profile and go through all regions to get all amis in the account
func GetAccountImages(account utils.AccountInfo) (AccountImages, error) {
	profile := account.Profile
	fmt.Println("Getting images for profile:", profile)
	imagesChan := make(chan RegionImages)
	var wg sync.WaitGroup

	for _, region := range utils.RegionMap {
		wg.Add(1)
		go func(region string) {
			var info RegionImages
			var err error
			defer wg.Done()
			account.Region = region
			sess, err := utils.GetSession(account)
			if err != nil {
				log.Println("Could not get users for", account.Profile, ":", err)
				return
			}
			info.Images, err = GetRegionImages(sess)
			if err != nil {
				log.Println("Could not get images for", region, "in", profile, ":", err)
				return
			}
			info.AccountID, err = utils.GetAccountID(sess)
			if err != nil {
				log.Println("Could not get account id for", account.Profile, ":", err)
				return
			}
			info.Region = region
			info.Profile = profile
			imagesChan <- info
		}(region)
	}

	go func() {
		wg.Wait()
		close(imagesChan)
	}()

	var accountImages AccountImages
	for regionImages := range imagesChan {
		accountImages = append(accountImages, regionImages)
	}

	return accountImages, nil
}

//GetProfilesImages will return all the images in all accounts of a given filename with a list of profiles in it
func GetProfilesImages(accounts []utils.AccountInfo) (ProfilesImages, error) {
	profilesImagesChan := make(chan AccountImages)
	var wg sync.WaitGroup

	for _, account := range accounts {
		wg.Add(1)
		go func(account utils.AccountInfo) {
			var err error
			defer wg.Done()
			accountImages, err := GetAccountImages(account)
			if err != nil {
				log.Println("Could not get images for", account.Profile, ":", err)
				return
			}
			profilesImagesChan <- accountImages
		}(account)
	}

	go func() {
		wg.Wait()
		close(profilesImagesChan)
	}()

	var profilesImages ProfilesImages
	for accountImages := range profilesImagesChan {
		profilesImages = append(profilesImages, accountImages)
	}
	return profilesImages, nil
}

func WriteProfilesImages(profileImages ProfilesImages, options utils.Ec2Options) error {
	outfile, err := utils.CreateFile("images.csv")
	if err != nil {
		return fmt.Errorf("could not create images file", err)
	}

	writer := csv.NewWriter(outfile)
	defer writer.Flush()
	fmt.Println("Writing images to file:", outfile.Name())

	var columnTitles = []string{"Profile",
		"Account ID",
		"Region",
		"Image Name",
		"Image ID",
		"Creation Time",
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
	for _, accountImages := range profileImages {
		for _, regionImages := range accountImages {
			for _, image := range regionImages.Images {
				var imageName string
				for _, tag := range image.Tags {
					if *tag.Key == "Name" {
						imageName = *tag.Value
					}
				}

				splitDate := strings.Split(*image.CreationDate, "T")
				startDate := splitDate[0]

				var data = []string{regionImages.Profile,
					regionImages.AccountID,
					regionImages.Region,
					imageName,
					*image.ImageId,
					startDate,
				}

				if len(tags) > 0 {
					for _, tag := range tags {
						x := false
						for _, imageTag := range image.Tags {
							if *imageTag.Key == tag {
								data = append(data, *imageTag.Value)
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
