package s3

import (
	"encoding/csv"
	"fmt"
	"strings"
	"strconv"
	"sync"

	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

/*
This file is for the functions to get the size of a bucket on a per object type basis.
Will show .jpg, .txt, .csv, etc and have a size per type and a total size for the bucket.
*/

type (
	BucketSizeInfo struct {
		BucketInfo  BucketInfo
		FileTypes   []FileType
		ObjectCount int
		TotalSize   int64
	}
	FileType struct {
		Type string
		Size int64
	}
)

//var ValidFileTypes = []string{"json", "yaml", "yml", "log", "tfstate", "csv"}
var ValidFileTypes = []string{"7z", "abc", "accdb", "apk", "bat", "bin", "bz2", "bzip2", "c", "c#", "cab", "cc", "cer", "cpp", "csv", "cxx", "dbf", "dbx", "deb", "dmg", "doc", "docx", "dot", "dotx", "dwg", "dxf", "eml", "emlx", "exe", "gpg", "gz", "gzip", "html", "iwa", "jar", "java", "json", "key", "keynote", "lua", "mdb", "msg", "msi", "odp", "oos", "p12", "pages", "pdf", "perl", "pgp", "pl", "pot", "pps", "ppt", "pptx", "pst", "py", "rar", "rtf", "sdp", "sdw", "sldasm", "slddrw", "sldprt", "sql", "sxi", "sxw", "tar.gz", "tsv", "txt", "vdx", "vsd", "vss", "vst", "vsx", "vtw", "vtx", "xls", "xlsx", "xlw", "xml", "xps", "zip"}

func GetBucketFileSize(bucket BucketInfo, sess *session.Session) (*BucketSizeInfo, error) {
	svc := s3.New(sess)
	params := &s3.ListObjectsInput{
		Bucket: aws.String(bucket.Name),
	}
	var bucketSizeInfo BucketSizeInfo
	bucketSizeInfo.BucketInfo = bucket
	x := false
	for !x {
		resp, err := svc.ListObjects(params)
		if err != nil {
			return nil, err
		}
		for _, object := range resp.Contents {
			//check if the key contains a period, which would mean it has a filetype
			// /path/to/file vs /path/to/file.txt
			var objectFileType string
			if strings.Contains(*object.Key, ".") {
				splitKey := strings.Split(*object.Key, ".")
				objectFileType = splitKey[len(splitKey)-1]
				valid := false
				for _, validFileType := range ValidFileTypes {
					//Checking for a set list of file types will limit the amount of different file types in the output
					//Otherwise, there could be hundreds which makes the report unusable
					if objectFileType == validFileType {
						valid = true
						break
					}
				}
				if !valid {
					objectFileType = "notValid"
				}
			} else {
				objectFileType = "notValid"
			}
			//check if filetype has already been counted
			found := false
			for i, fileType := range bucketSizeInfo.FileTypes {
				if fileType.Type == objectFileType {
					found = true
					bucketSizeInfo.FileTypes[i].Size += *object.Size
				}
			}
			//if not found, add to known file types and map
			if !found {
				newFileType := FileType{
					Type: objectFileType,
					Size: *object.Size,
				}
				bucketSizeInfo.FileTypes = append(bucketSizeInfo.FileTypes, newFileType)
			}
			//increment total size, and object count
			bucketSizeInfo.TotalSize += *object.Size
			bucketSizeInfo.ObjectCount += 1
		}

		if resp.NextMarker != nil {
			params.Marker = resp.NextMarker
		} else {
			break
		}
	}
	return &bucketSizeInfo, nil
}

func GetProfileBucketsFileSize(buckets []BucketInfo, account utils.AccountInfo) ([]*BucketSizeInfo, error) {
	getBucketsChan := make(chan *BucketSizeInfo)
	var wg sync.WaitGroup
	sess, err := account.GetSession()
	if err != nil {
		utils.LogAll("could not get sess for ", account.Profile, ":", err)
	}
	accountId, err := utils.GetAccountId(sess)
	if err != nil {
		utils.LogAll("could not get account id for", account.Profile, ":", err)
	}
	for _, bucket := range buckets {
		wg.Add(1)
		go func(bucket BucketInfo) {
			var err error
			defer wg.Done()
			sess, err := account.GetSession()
			if err != nil {
				utils.LogAll("could not open session for", account.Profile, ":", err)
				return
			}
			//Need to get the bucket region in the session in order to get the contents
			bucketRegion, err := GetBucketRegion(sess, bucket.Name)
			if err != nil {
				utils.LogAll("could not get the region for", bucket.Name, "in account", account.Profile, ":", err)
				return
			}
			account.Region = bucketRegion
			sess, err = account.GetSession()
			bucketSizeInfo, err := GetBucketFileSize(bucket, sess)
			if err != nil {
				utils.LogAll("could not get bucketinfo for", account.Profile, ":", err)
				return
			}
			bucketSizeInfo.BucketInfo.Region = bucketRegion
			bucketSizeInfo.BucketInfo.Profile = account.Profile
			bucketSizeInfo.BucketInfo.AccountId = accountId

			getBucketsChan <- bucketSizeInfo
		}(bucket)
	}
	go func() {
		wg.Wait()
		close(getBucketsChan)
	}()

	var bucketsSizeInfo []*BucketSizeInfo
	for bucketSizeInfo := range getBucketsChan {
		bucketsSizeInfo = append(bucketsSizeInfo, bucketSizeInfo)
	}
	return bucketsSizeInfo, nil
}

func GetProfilesPublicBucketsFileSize(accounts []utils.AccountInfo, bucketOption string) ([]*BucketSizeInfo, error) {
	getBucketsChan := make(chan *BucketSizeInfo)
	var wg sync.WaitGroup

	for _, account := range accounts {
		wg.Add(1)
		go func(account utils.AccountInfo) {
			defer wg.Done()
			buckets, err := GetProfileBuckets(account, "")
			if err != nil {
				utils.LogAll("could not get buckets for profile", account.Profile, ":", err)
				return
			}
			if bucketOption == "public-only" {
				var publicBuckets []BucketInfo
				if err != nil {
					utils.LogAll("could not public bucket check session:", err)
					return
				}
				for _, bucket := range buckets {
					account.Region = bucket.Region
					sess, err := account.GetSession()
					public, err := CheckPublicBucket(bucket.Name, sess)
					if err != nil {
						utils.LogAll("could not check public bucket", bucket.Name, ":", err)
						continue
					}
					if public {
						publicBuckets = append(publicBuckets, bucket)
					}
				}
				bucketsSizeInfo, err := GetProfileBucketsFileSize(publicBuckets, account)
				if err != nil {
					utils.LogAll("could not get bucket info for profile", account.Profile, ":", err)
				}

				for _, bucketSizeInfo := range bucketsSizeInfo {
					getBucketsChan <- bucketSizeInfo
				}
			}else {
				bucketsSizeInfo, err := GetProfileBucketsFileSize(buckets, account)
				if err != nil {
					utils.LogAll("could not get bucket info for profile", account.Profile, ":", err)
				}

				for _, bucketSizeInfo := range bucketsSizeInfo {
					getBucketsChan <- bucketSizeInfo
				}
			}
		}(account)
	}
	go func() {
		wg.Wait()
		close(getBucketsChan)
	}()

	var bucketsSizeInfo []*BucketSizeInfo
	for bucketSizeInfo := range getBucketsChan {
		bucketsSizeInfo = append(bucketsSizeInfo, bucketSizeInfo)
	}
	return bucketsSizeInfo, nil
}

func WriteProfilesBucketsFileSize(profilesBuckets []*BucketSizeInfo) error {
	outputDir := "output/s3/"
	utils.MakeDir(outputDir)
	outputFile := outputDir + "bucketsSize.csv"
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
		"Object Count",
	}

	var fileTypes []string
	for _, bucket := range profilesBuckets {
		for _, fileType := range bucket.FileTypes {
			x := false
			for _, column := range columnTitles {
				if column == fileType.Type {
					x = true
				}
			}
			if !x {
				fileTypes = append(fileTypes, fileType.Type)
				columnTitles = append(columnTitles, fileType.Type)
			}
		}
	}

	//tags := options.Tags
	//if len(tags) > 0 {
	//	for _, tag := range tags {
	//		columnTitles = append(columnTitles, tag)
	//	}
	//}

	columnTitles = append(columnTitles, "Total Size")
	err = writer.Write(columnTitles)
	if err != nil {
		fmt.Println(err)
	}
	for _, bucket := range profilesBuckets {
			var data = []string{bucket.BucketInfo.Profile,
				bucket.BucketInfo.AccountId,
				bucket.BucketInfo.Region,
				bucket.BucketInfo.Name,
				strconv.Itoa(bucket.ObjectCount),
			}

			for i, column := range columnTitles {
				//skip the profile, account id, region, name, and object count columns
				if i <= 4 || column == "Total Size"{
					continue
				}
				x := false
				for _, fileType := range bucket.FileTypes {
					if column == fileType.Type {
						x = true
						data = append(data, strconv.FormatInt(fileType.Size, 10))
					}
				}
				if !x {
					data = append(data, "N/A")
				}
			}
			data = append(data, strconv.FormatInt(bucket.TotalSize, 10))

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
	return nil
}
