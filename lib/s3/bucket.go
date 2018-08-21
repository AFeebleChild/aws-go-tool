package s3

import (
	"strings"

	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type BucketInfo struct {
	Name    string `yaml:"name"`
	Profile string `yaml:"profile"`
	Region  string `yaml:"region"`
	//Err is for tracking any errors with getting the bucket Region in the GetAllBucketsInfo function
	//Or even for use in later functions to track errors with this specific bucket
	Error string `yaml:"error"`
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

func GetAllBucketInfo(profile string, name string) ([]BucketInfo, error) {
	sess := utils.OpenSession(profile, "us-east-1")
	bucketNames, err := GetAllBucketNames(sess, name)
	if err != nil {
		return nil, err
	}

	var buckets []BucketInfo

	for _, bucketName := range bucketNames {
		var bucket BucketInfo
		bucket.Name = bucketName
		bucket.Profile = profile
		bucket.Error = ""
		bucket.Region, err = GetBucketRegion(sess, bucketName)
		if err != nil {
			bucket.Error = "could not get the bucket region"
		}
		buckets = append(buckets, bucket)
	}
	return buckets, nil
}

//GetAllBucketNames will return all of the bucket names for the listed account
//Can filter down the names with the "name" parameter if desired
func GetAllBucketNames(sess *session.Session, name string) ([]string, error) {
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
			if strings.Contains(*bucket.Name, "autopark") {
				names = append(names, *bucket.Name)
			}
		}
	}

	return names, nil
}
