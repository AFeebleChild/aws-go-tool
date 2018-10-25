package vpc

import (
	"encoding/csv"
	"fmt"
	"log"
	"strings"
	"strconv"
	"sync"

	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type RegionSubnets struct {
	AccountId   string
	Region      string
	Profile     string
	Subnets     []ec2.Subnet
	RouteTables []ec2.RouteTable
}

type AccountSubnets []RegionSubnets
type ProfilesSubnets []AccountSubnets

//GetRegionSubnets will take a session and pull all subnets and subnets based on the region of the session
func GetRegionSubnets(sess *session.Session) (*RegionSubnets, error) {
	var subnets RegionSubnets
	svc := ec2.New(sess)
	params := &ec2.DescribeSubnetsInput{}

	//resp, err := ec2.New(sess).DescribeSubnets(params)
	resp, err := svc.DescribeSubnets(params)
	if err != nil {
		return nil, err
	}

	for _, subnet := range resp.Subnets {
		subnets.Subnets = append(subnets.Subnets, *subnet)
	}

	rtparams := &ec2.DescribeRouteTablesInput{}

	rtresp, err := svc.DescribeRouteTables(rtparams)
	if err != nil {
		return nil, err
	}

	for _, routeTable := range rtresp.RouteTables {
		subnets.RouteTables = append(subnets.RouteTables, *routeTable)
	}

	return &subnets, nil
}

//GetAccountSubnets will take a profile and go through all regions to get all instances in the account
func GetAccountSubnets(account utils.AccountInfo) (AccountSubnets, error) {
	profile := account.Profile
	fmt.Println("Getting subnet info for profile:", profile)
	subnetsChan := make(chan RegionSubnets)
	var wg sync.WaitGroup

	for _, region := range utils.RegionMap {
		wg.Add(1)
		go func(region string) {
			defer wg.Done()

			var err error
			account.Region = region
			sess, err := utils.GetSession(account)
			if err != nil {
				log.Println("Could not get session for", account.Profile, ":", err)
				return
			}
			subnets, err := GetRegionSubnets(sess)
			if err != nil {
				log.Println("Could not get subnet info for", region, "in", profile, ":", err)
				return
			}
			subnets.AccountId, err = utils.GetAccountID(sess)
			if err != nil {
				log.Println("Could not get account id for", profile, ":", err)
				return
			}
			subnets.Region = region
			subnets.Profile = account.Profile
			subnetsChan <- *subnets
		}(region)
	}

	go func() {
		wg.Wait()
		close(subnetsChan)
	}()

	var accountSubnets AccountSubnets
	for regionSubnets := range subnetsChan {
		accountSubnets = append(accountSubnets, regionSubnets)
	}

	return accountSubnets, nil
}

//GetProfilesSubnets will return all the subnets in all accounts of a given filename with a list of profiles in it
func GetProfilesSubnets(accounts []utils.AccountInfo) (ProfilesSubnets, error) {
	profilesSubnetsChan := make(chan AccountSubnets)
	var wg sync.WaitGroup

	for _, account := range accounts {
		wg.Add(1)
		go func(account utils.AccountInfo) {
			defer wg.Done()

			var err error
			accountSubnets, err := GetAccountSubnets(account)
			if err != nil {
				log.Println("Could not get subnet info for", account.Profile, ":", err)
				return
			}
			profilesSubnetsChan <- accountSubnets
		}(account)
	}

	go func() {
		wg.Wait()
		close(profilesSubnetsChan)
	}()

	var profilesSubnets ProfilesSubnets
	for accountSubnets := range profilesSubnetsChan {
		profilesSubnets = append(profilesSubnets, accountSubnets)
	}
	return profilesSubnets, nil
}

//CheckPublicSubnet will check whether the route table associated with a subnet is actually public or not
//returns true if public, false if not
//TODO this is not the best check relying on the name.  Need to add a route check on the route table for an IGW
func CheckPublicSubnet(subnetId string, routeTables []ec2.RouteTable) (bool, error) {
	//loop through route tables
	for _, routeTable := range routeTables {
		//loop through subnet associations
		for _, association := range routeTable.Associations {
			//compare the association subnet id to the provided subnet id
			if association.SubnetId == nil {
				continue
			}else if *association.SubnetId == subnetId {
				//if it matches, start looping through tags to check the name
				for _, tag := range routeTable.Tags {
					//set value to lower to have fewer punctuation issues
					lowerValue := strings.ToLower(*tag.Value)
					if *tag.Key == "Name" && strings.Contains(lowerValue, "public") {
						return true, nil
					}
				}
			}
		}
	}
	return false, nil
}

func WriteProfilesSubnets(profileSubnets ProfilesSubnets) error{
	outputDir := "output/vpc/"
	utils.MakeDir(outputDir)
	outputFile := outputDir + "subnets.csv"
	outfile, err := utils.CreateFile(outputFile)
	if err != nil {
		return fmt.Errorf("could not create subnets file", err)
	}

	writer := csv.NewWriter(outfile)
	defer writer.Flush()
	fmt.Println("Writing subnets to file:", outfile.Name())

	var columnTitles = []string{"Account",
		"Account ID",
		"Region",
		"Resource Name",
		"Subnet ID",
		"VPC ID",
		"CIDR Block",
		"Is Default",
		"Public Route Table Validated",
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

	for _, accountSubnets := range profileSubnets {
		for _, regionSubnets := range accountSubnets {
			for _, subnet := range regionSubnets.Subnets {
				var subnetName string
				for _, tag := range subnet.Tags {
					if *tag.Key == "Name" {
						subnetName = *tag.Value
					}
				}

				lowerName := strings.ToLower(subnetName)
				validPublicSubnet := "N/A"
				if strings.Contains(lowerName, "public") {
					x, _ := CheckPublicSubnet(*subnet.SubnetId, regionSubnets.RouteTables)
					validPublicSubnet = strconv.FormatBool(x)
				}

				var data = []string{regionSubnets.Profile,
					regionSubnets.AccountId,
					regionSubnets.Region,
					subnetName,
					*subnet.SubnetId,
					*subnet.VpcId,
					*subnet.CidrBlock,
					strconv.FormatBool(*subnet.DefaultForAz),
					validPublicSubnet,
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
