package cmd

import (
	"fmt"

	"github.com/afeeblechild/aws-go-tool/lib/s3"
	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/spf13/cobra"
)

var s3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "For use with interacting with the s3 service",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Run -h to see the help menu")
	},
}

var bucketsListCmd = &cobra.Command{
	Use:   "bucketslist",
	Short: "Will generate a report of bucket info for all given accounts",
	Run: func(cmd *cobra.Command, args []string) {
		profilesBuckets, err := s3.GetProfilesBuckets(Accounts)
		if err != nil {
			utils.LogAll("could not get buckets:", err)
			return
		}
		// TODO add tag support
		//options := utils.Ec2Options{Tags:Tags}
		err = s3.WriteProfilesBuckets(profilesBuckets)
		if err != nil {
			utils.LogAll("could not write buckets:", err)
			return
		}
	},
}

var fileSizeCmd = &cobra.Command{
	Use:   "filesize",
	Short: "To get the size of objects in buckets per object type",
	Run: func(cmd *cobra.Command, args []string) {
		if BucketFile == "public-only" {
			bucketsInfo, err := s3.GetProfilesPublicBucketsFileSize(Accounts, "public-only")

			if err != nil {
				utils.LogAll("could not get profiles buckets:", err)
				return
			}
			s3.WriteProfilesBucketsFileSize(bucketsInfo)
		} else if BucketFile == "all" {
			bucketsInfo, err := s3.GetProfilesPublicBucketsFileSize(Accounts, "public-only")

			if err != nil {
				utils.LogAll("could not get profiles buckets:", err)
				return
			}
			s3.WriteProfilesBucketsFileSize(bucketsInfo)
		} else {
			buckets, err := utils.ReadFile(BucketFile)
			if err != nil {
				utils.LogAll("could not read buckets file:", err)
				return
			}

			var bucketsInfo []s3.BucketInfo
			for _, bucket := range buckets {
				bucketInfo := s3.BucketInfo{Name: bucket}
				bucketsInfo = append(bucketsInfo, bucketInfo)
			}

			bucketsSizeInfo, err := s3.GetProfileBucketsFileSize(bucketsInfo, Accounts[0])
			if err != nil {
				utils.LogAll("could not get buckets size info:", err)
				return
			}
			utils.PrettyPrintJson(bucketsSizeInfo)
		}
		//for _, bucketInfo := range bucketsInfo {
		//	fmt.Println("BucketName:", bucketInfo.BucketName)
		//	for _, fileType := range bucketInfo.FileTypes {
		//		fmt.Printf("%s: %v\n", fileType.Type, fileType.Size)
		//	}
		//	fmt.Println("ObjectCount:", bucketInfo.ObjectCount)
		//	fmt.Println("TotalSize:", bucketInfo.TotalSize)
		//}
	},
}

var BucketFile string

func init() {
	RootCmd.AddCommand(s3Cmd)

	s3Cmd.AddCommand(bucketsListCmd)
	s3Cmd.AddCommand(fileSizeCmd)

	s3Cmd.PersistentFlags().StringVarP(&BucketFile, "bucketfile", "b", "", "file with list of buckets")
}
