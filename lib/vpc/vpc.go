package vpc

import (
	"fmt"
	"strings"
	"sync"

	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type RegionNetworks struct {
	Region  string
	Profile string
	Vpcs    []ec2.Vpc
	Subnets []ec2.Subnet
}

type AccountNetworks []RegionNetworks
type ProfilesNetworks []AccountNetworks

//GetRegionNetworks will take a session and pull all vpcs and subnets based on the region of the session
func GetRegionNetworks(sess *session.Session, arn string) (*RegionNetworks, error) {
	//TODO assume role stuff here
	creds := stscreds.NewCredentials(sess, arn)
	svc := ec2.New(sess, &aws.Config{Credentials: creds})
	var networks RegionNetworks
	vpcParams := &ec2.DescribeVpcsInput{}
	subnetParams := &ec2.DescribeSubnetsInput{}

	vpcResp, err := svc.DescribeVpcs(vpcParams)
	if err != nil {
		return nil, err
	}
	subnetResp, err := svc.DescribeSubnets(subnetParams)
	if err != nil {
		return nil, err
	}

	for _, vpc := range vpcResp.Vpcs {
		networks.Vpcs = append(networks.Vpcs, *vpc)
	}
	for _, subnet := range subnetResp.Subnets {
		networks.Subnets = append(networks.Subnets, *subnet)
	}

	return &networks, nil
}

//GetAccountNetworks will take a profile and go through all regions to get all instances in the account
func GetAccountNetworks(profile string) (AccountNetworks, error) {
	fmt.Println("Getting vpcs/subnets for profile:", profile)
	networksChan := make(chan RegionNetworks)
	var wg sync.WaitGroup

	for _, region := range utils.RegionMap {
		wg.Add(1)
		go func(region string) {
			defer wg.Done()

			var err error
			sess := utils.OpenSession(profile, region)
			networks, err := GetRegionNetworks(sess, profile)
			if err != nil {
				fmt.Fprintln(logFile, "Could not get vpcs/subnets for", region, "in", profile, ":", err)
				return
			}
			networks.Region = region
			//TODO assume role stuff here
			splitProfile := strings.Split(profile, ":")
			networks.Profile = splitProfile[4]
			networksChan <- *networks
		}(region)
	}

	go func() {
		wg.Wait()
		close(networksChan)
	}()

	var accountNetworks AccountNetworks
	for regionNetworks := range networksChan {
		accountNetworks = append(accountNetworks, regionNetworks)
	}

	return accountNetworks, nil
}

//GetProfilesNetworks will return all the vpcs/subnets in all accounts of a given filename with a list of profiles in it
func GetProfilesNetworks(filename string) (ProfilesNetworks, error) {
	//TODO add cross account role support
	profiles, err := utils.ReadProfilesFile(filename)
	if err != nil {
		return nil, err
	}
	profilesNetworksChan := make(chan AccountNetworks)
	var wg sync.WaitGroup

	for _, profile := range profiles {
		wg.Add(1)
		go func(profile string) {
			defer wg.Done()

			var err error
			accountNetworks, err := GetAccountNetworks(profile)
			if err != nil {
				fmt.Fprintln(logFile, "Could not get vpcs/subnets for", profile, ":", err)
				return
			}
			profilesNetworksChan <- accountNetworks
		}(profile)
	}

	go func() {
		wg.Wait()
		close(profilesNetworksChan)
	}()

	var profilesNetworks ProfilesNetworks
	for accountnetworks := range profilesNetworksChan {
		profilesNetworks = append(profilesNetworks, accountnetworks)
	}
	return profilesNetworks, nil
}

func WriteProfilesNetworks(profileNetworks ProfilesNetworks) {
	outfile, err := utils.CreateFile("Networks.csv")
	fmt.Println("Writing vpcs/subnets to file:", outfile.Name())
	if err != nil {
		fmt.Println("Could not open outfile to write info")
		panic(err)
	}

	fmt.Fprintf(outfile, "Account, Region, Resource Name, VPC ID, Subnet ID, CIDR Block, Is Default\n")
	for _, accountNetworks := range profileNetworks {
		for _, regionNetworks := range accountNetworks {
			for _, vpc := range regionNetworks.Vpcs {
				var vpcName string
				for _, tag := range vpc.Tags {
					if *tag.Key == "Name" {
						vpcName = *tag.Value
					}
				}

				fmt.Fprintf(outfile, "%s, %s, %s, %s, %s, %s, %t\n", regionNetworks.Profile, regionNetworks.Region, vpcName, *vpc.VpcId, "N/A", *vpc.CidrBlock, *vpc.IsDefault)
			}
			for _, subnet := range regionNetworks.Subnets {
				var subnetName string
				for _, tag := range subnet.Tags {
					if *tag.Key == "Name" {
						subnetName = *tag.Value
					}
				}

				fmt.Fprintf(outfile, "%s, %s, %s, %s, %s, %s, %t\n", regionNetworks.Profile, regionNetworks.Region, subnetName, *subnet.VpcId, *subnet.SubnetId, *subnet.CidrBlock, *subnet.DefaultForAz)
			}
		}
	}
}
