package vpc

import (
	"encoding/csv"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type RegionVpcs struct {
	AccountId string
	Region    string
	Profile   string
	Vpcs      []ec2.Vpc
	Subnets   []ec2.Subnet
}

type AccountVpcs []RegionVpcs
type ProfilesVpcs []AccountVpcs

//GetRegionVpcs will take a session and pull all vpcs and subnets based on the region of the session
func GetRegionVpcs(sess *session.Session, arn string) (*RegionVpcs, error) {
	svc := ec2.New(sess)
	var vpcs RegionVpcs
	params := &ec2.DescribeVpcsInput{}

	vpcResp, err := svc.DescribeVpcs(params)
	if err != nil {
		return nil, err
	}

	for _, vpc := range vpcResp.Vpcs {
		vpcs.Vpcs = append(vpcs.Vpcs, *vpc)
	}
	subnets, err := GetRegionSubnets(sess)
	if err != nil {
		return nil, err
	}

	vpcs.Subnets = subnets.Subnets

	return &vpcs, nil
}

//GetAccountVpcs will take a profile and go through all regions to get all instances in the account
func GetAccountVpcs(account utils.AccountInfo) (AccountVpcs, error) {
	profile := account.Profile
	fmt.Println("Getting vpc info for profile:", profile)
	vpcsChan := make(chan RegionVpcs)
	var wg sync.WaitGroup

	for _, region := range utils.RegionMap {
		wg.Add(1)
		go func(region string) {
			defer wg.Done()

			var err error
			account.Region = region
			sess, err := account.GetSession()
			if err != nil {
				log.Println("Could not get session for", account.Profile, ":", err)
				return
			}
			vpcs, err := GetRegionVpcs(sess, profile)
			if err != nil {
				log.Println("Could not get vpc info for", region, "in", profile, ":", err)
				return
			}
			vpcs.AccountId, err = utils.GetAccountId(sess)
			if err != nil {
				log.Println("Could not get account id for", profile, ":", err)
				return
			}
			vpcs.Region = region
			vpcs.Profile = account.Profile
			vpcsChan <- *vpcs
		}(region)
	}

	go func() {
		wg.Wait()
		close(vpcsChan)
	}()

	var accountVpcs AccountVpcs
	for regionVpcs := range vpcsChan {
		accountVpcs = append(accountVpcs, regionVpcs)
	}

	return accountVpcs, nil
}

//GetProfilesVpcs will return all the vpcs/subnets in all accounts of a given filename with a list of profiles in it
func GetProfilesVpcs(accounts []utils.AccountInfo) (ProfilesVpcs, error) {
	profilesVpcsChan := make(chan AccountVpcs)
	var wg sync.WaitGroup

	for _, account := range accounts {
		wg.Add(1)
		go func(account utils.AccountInfo) {
			defer wg.Done()

			var err error
			accountVpcs, err := GetAccountVpcs(account)
			if err != nil {
				log.Println("Could not get vpc info for", account.Profile, ":", err)
				return
			}
			profilesVpcsChan <- accountVpcs
		}(account)
	}

	go func() {
		wg.Wait()
		close(profilesVpcsChan)
	}()

	var profilesVpcs ProfilesVpcs
	for accountVpcs := range profilesVpcsChan {
		profilesVpcs = append(profilesVpcs, accountVpcs)
	}
	return profilesVpcs, nil
}

func WriteProfilesVpcs(profileVpcs ProfilesVpcs) error {
	outputDir := "output/vpc/"
	utils.MakeDir(outputDir)
	outputFile := outputDir + "vpcs.csv"
	outfile, err := utils.CreateFile(outputFile)
	if err != nil {
		return fmt.Errorf("could not create vpcs file", err)
	}

	writer := csv.NewWriter(outfile)
	defer writer.Flush()
	fmt.Println("Writing vpcs to file:", outfile.Name())

	var columnTitles = []string{"Account",
		"Account ID",
		"Region",
		"Resource Name",
		"VPC ID",
		"Subnet ID",
		"CIDR Block",
		"Is Default",
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

	for _, accountVpcs := range profileVpcs {
		for _, regionVpcs := range accountVpcs {
			for _, vpc := range regionVpcs.Vpcs {
				var vpcName string
				for _, tag := range vpc.Tags {
					if *tag.Key == "Name" {
						vpcName = *tag.Value
					}
				}

				var data = []string{regionVpcs.Profile,
					regionVpcs.AccountId,
					regionVpcs.Region,
					vpcName,
					*vpc.VpcId,
					"N/A",
					*vpc.CidrBlock,
					strconv.FormatBool(*vpc.IsDefault),
				}

				//if len(tags) > 0 {
				//	for _, tag := range tags {
				//		x := false
				//		for _, imageTag := range image.Tags {
				//			if *imageTag.Key == tag {
				//				data = append(data, *imageTag.Value)
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
			for _, subnet := range regionVpcs.Subnets {
				var subnetName string
				for _, tag := range subnet.Tags {
					if *tag.Key == "Name" {
						subnetName = *tag.Value
					}
				}

				var data = []string{regionVpcs.Profile,
					regionVpcs.AccountId,
					regionVpcs.Region,
					subnetName,
					*subnet.VpcId,
					*subnet.SubnetId,
					*subnet.CidrBlock,
					strconv.FormatBool(*subnet.DefaultForAz),
				}

				//if len(tags) > 0 {
				//	for _, tag := range tags {
				//		x := false
				//		for _, imageTag := range image.Tags {
				//			if *imageTag.Key == tag {
				//				data = append(data, *imageTag.Value)
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
