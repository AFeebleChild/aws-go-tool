package s3

import (
	"strings"

	"encoding/csv"
	"fmt"
	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"sync"
)

type BucketInfo struct {
	Name      string `yaml:"name"`
	Profile   string `yaml:"profile"`
	AccountId string
	Region    string `yaml:"region"`
	//Err is for tracking any errors with getting the bucket Region in the GetAllBucketsInfo function
	//Or even for use in later functions to track errors with this specific bucket
	Error error `yaml:"error"`
}

type AccountBuckets []BucketInfo
type ProfilesBuckets []AccountBuckets

func GetProfilesBuckets(accounts []utils.AccountInfo) (ProfilesBuckets, error){
	profilesBucketsChan := make(chan AccountBuckets)
	var wg sync.WaitGroup

	for _, account := range accounts {
		wg.Add(1)
		go func(account utils.AccountInfo) {
			var err error
			defer wg.Done()
			accountBuckets, err := GetAccountBuckets(account.Profile, "")
			if err != nil {
				log.Println("Could not get buckets for", account.Profile, ":", err)
				return
			}
			profilesBucketsChan <- accountBuckets
		}(account)
	}

	go func() {
		wg.Wait()
		close(profilesBucketsChan)
	}()

	var profilesBuckets ProfilesBuckets
	for accountBuckets := range profilesBucketsChan {
		profilesBuckets = append(profilesBuckets, accountBuckets)
	}
	return profilesBuckets, nil
}

//name is an optional parameter for searching the buckets for a name
func GetAccountBuckets(profile string, name string) ([]BucketInfo, error) {
	fmt.Println("Getting buckets for profile:", profile)
	//region is arbitrary for S3 buckets, since the names can be retrieved from any region
	sess := utils.OpenSession(profile, "us-east-1")
	bucketNames, err := GetBucketNames(sess, name)
	if err != nil {
		return nil, err
	}

	var buckets []BucketInfo

	accountId, err := utils.GetAccountId(sess)
	if err != nil {
		log.Println("could not get account ID for profile: ", profile)
	}
	for _, bucketName := range bucketNames {
		var bucket BucketInfo
		bucket.Name = bucketName
		bucket.Profile = profile
		bucket.AccountId = accountId
		bucket.Error = nil
		bucket.Region, err = GetBucketRegion(sess, bucketName)
		if err != nil {
			bucket.Error = fmt.Errorf("could not get the bucket region :", err)
		}
		buckets = append(buckets, bucket)
	}
	return buckets, nil
}

//GetAllBucketNames will return all of the bucket names for the listed account
//Can filter down the names with the "name" parameter if desired
func GetBucketNames(sess *session.Session, name string) ([]string, error) {
	params := &s3.ListBucketsInput{}

	resp, err := s3.New(sess).ListBuckets(params)
	if err != nil {
		return nil, err
	}

	var names []string
	if name == "" {
		for _, bucket := range resp.Buckets {
			names = append(names, *bucket.Name)
		}
	} else {
		for _, bucket := range resp.Buckets {
			if strings.Contains(*bucket.Name, name) {
				names = append(names, *bucket.Name)
			}
		}
	}

	return names, nil
}

func GetBucketRegion(sess *session.Session, bucketName string) (string, error) {
	params := &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketName),
	}

	resp, err := s3.New(sess).GetBucketLocation(params)
	if err != nil {
		return "", err
	}

	//if the Location is nil, the region is built
	if resp.LocationConstraint == nil {
		return "us-east-1", nil
	}
	return *resp.LocationConstraint, nil
}

func WriteProfilesBuckets(profileBuckets ProfilesBuckets) error {
	outputDir := "output/s3/"
	utils.MakeDir(outputDir)
	outputFile := outputDir + "buckets.csv"
	outfile, err := utils.CreateFile(outputFile)
	if err != nil {
		return fmt.Errorf("could not create buckets file", err)
	}

	writer := csv.NewWriter(outfile)
	defer writer.Flush()
	fmt.Println("Writing buckets to file:", outfile.Name())
	var columnTitles = []string{"Profile",
		"Account ID",
		"Region",
		"Bucket Name",
	}

	//tags := options.Tags
	//if len(tags) > 0 {
	//	for _, tag := range tags {
	//		columnTitles = append(columnTitles, tag)
	//	}
	//}

	err = writer.Write(columnTitles)
	if err != nil {
		fmt.Println(err)
	}
	//var pemKeys []string
	//pemKeyFile, _ := utils.CreateFile("pemKeys.csv")
	for _, accountBuckets := range profileBuckets {
		for _, bucket := range accountBuckets {
			var data = []string{bucket.Profile,
				bucket.AccountId,
				bucket.Region,
				bucket.Name,
			}

			//if len(tags) > 0 {
			//	for _, tag := range tags {
			//		x := false
			//		for _, bucketTag := range bucket.Tags {
			//			if *bucketTag.Key == tag {
			//				data = append(data, *bucketTag.Value)
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
	return nil
}
