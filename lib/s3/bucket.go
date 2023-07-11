package s3

import (
	"encoding/csv"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type BucketInfo struct {
	Name       string `yaml:"name"`
	Profile    string `yaml:"profile"`
	AccountId  string
	Encryption string `yaml:"encrypted"`
	Region     string `yaml:"region"`
}

type AccountBuckets []BucketInfo
type ProfilesBuckets []AccountBuckets

//Will return true if the bucket is public
func CheckPublicBucket(bucketName string, sess *session.Session) (bool, error) {
	params := &s3.GetBucketAclInput{
		Bucket: aws.String(bucketName),
	}

	resp, err := s3.New(sess).GetBucketAcl(params)
	if err != nil {
		return false, err
	}

	for _, grant := range resp.Grants {
		//Canonical user type is the root user of the bucket
		//If any other user has access, it should be a public bucket
		if *grant.Grantee.Type != "CanonicalUser" && *grant.Grantee.URI != "http://acs.amazonaws.com/groups/s3/LogDelivery" {
			//fmt.Println("Bucket:", bucketName, ": Grantee -", grant.Grantee, ": Permission -", *grant.Permission)
			return true, nil
		}
	}

	return false, nil
}

func GetProfilesBuckets(accounts []utils.AccountInfo) (ProfilesBuckets, error) {
	profilesBucketsChan := make(chan AccountBuckets)
	var wg sync.WaitGroup

	for _, account := range accounts {
		wg.Add(1)
		go func(account utils.AccountInfo) {
			var err error
			defer wg.Done()
			accountBuckets, err := GetProfileBuckets(account, "")
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
func GetProfileBuckets(account utils.AccountInfo, name string) ([]BucketInfo, error) {
	profile := account.Profile
	fmt.Println("Getting buckets for profile:", profile)
	sess, err := account.GetSession()
	if err != nil {
		return nil, err
	}
	bucketNames, err := GetBucketNames(sess, name)
	if err != nil {
		return nil, err
	}

	var buckets []BucketInfo

	accountId, err := utils.GetAccountId(sess)
	if err != nil {
		utils.LogAll("could not get account ID for profile: ", profile)
	}
	for _, bucketName := range bucketNames {
		var bucket BucketInfo
		bucket.Name = bucketName
		bucket.Profile = profile
		bucket.AccountId = accountId
		bucket.Region, err = GetBucketRegion(sess, bucketName)
		if err != nil {
			utils.LogAll("could not get region for", bucketName, "in", profile, ":", err)
		}
		encryptionSess := sess
		//Region matters for getting the encryption information, so just create a temp account in order to set the proper region and get a new session
		if bucket.Region != "us-east-1" {
			tempAccount := account
			tempAccount.Region = bucket.Region
			encryptionSess, err = tempAccount.GetSession()
			//TODO determine a good fail condition here
			if err != nil {

			}
		}
		bucket.Encryption, err = GetBucketEncryption(encryptionSess, bucketName)
		if err != nil {
			utils.LogAll("could not get encryption for", bucketName, "in", profile, ":", err)
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

//GetBucketEncryption will return the encryption type, if enabled, for the specified bucket
//There should only ever be one encrypt type applied, even though the rules is a slice
func GetBucketEncryption(sess *session.Session, name string) (string, error) {
	params := &s3.GetBucketEncryptionInput{
		Bucket: aws.String(name),
	}

	resp, err := s3.New(sess).GetBucketEncryption(params)
	if err != nil {
		if strings.Contains(err.Error(), "The server side encryption configuration was not found") {
			return "no encryption", nil
		}
		return "", err
	}

	//Just getting down to the encryption type to return, and nothing else
	return *resp.ServerSideEncryptionConfiguration.Rules[0].ApplyServerSideEncryptionByDefault.SSEAlgorithm, nil
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
		return fmt.Errorf("could not create buckets file: %v", err)
	}

	writer := csv.NewWriter(outfile)
	defer writer.Flush()
	fmt.Println("Writing buckets to file:", outfile.Name())
	var columnTitles = []string{"Profile",
		"Account ID",
		"Region",
		"Bucket Name",
		"Encryption",
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
				bucket.Encryption,
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
