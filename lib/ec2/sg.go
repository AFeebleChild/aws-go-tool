package ec2

import (
	"fmt"
	"os"
	"sync"
	"log"
	"encoding/csv"

	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type SGOptions struct {
	Cidr string
	Tags []string
}

type RegionSGs struct {
	Region  string
	Profile string
	AccountID string
	SGs     []ec2.SecurityGroup
}

type AccountSGs []RegionSGs
type ProfilesSGs []AccountSGs

func GetSGGroupRules(SGID string, file *os.File, sess *session.Session) error {
	params := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("group-id"),
				Values: []*string{
					aws.String(SGID),
				},
			},
		},
	}

	resp, err := ec2.New(sess).DescribeSecurityGroups(params)
	if err != nil {
		return err
	}

	SG := resp.SecurityGroups[0]

	for _, rule := range SG.IpPermissions {
		for _, ip := range rule.IpRanges {
			fmt.Fprintf(file, "%s, %s, %s, %s, %s, %s, %s\n", SGID, *SG.GroupName, *rule.IpProtocol, *ip.CidrIp, *rule.FromPort, *rule.ToPort, *ip.CidrIp)
		}
	}

	return nil
}

func GetRegionSGs(sess *session.Session) ([]ec2.SecurityGroup, error) {
	params := &ec2.DescribeSecurityGroupsInput{
		DryRun: aws.Bool(false),
	}

	var SGs []ec2.SecurityGroup
	//x is the check to ensure there is no roles left from the IsTruncated
	x := true
	for x {
		resp, err := ec2.New(sess).DescribeSecurityGroups(params)
		if err != nil {
			return nil, err
		}
		for _, SG := range resp.SecurityGroups {
			SGs = append(SGs, *SG)
		}
		//If it is truncated, add the marker to the params for the next loop
		//If not, set x to false to exit for loop
		if resp.NextToken != nil {
			params.NextToken = resp.NextToken
		} else {
			x = false
		}
	}

	return SGs, nil
}

func GetAccountSGs(account utils.AccountInfo) (AccountSGs, error) {
	profile := account.Profile
	fmt.Println("Getting Security Groups for profile:", profile)
	SGsChan := make(chan RegionSGs)
	var wg sync.WaitGroup

	for _, region := range utils.RegionMap {
		wg.Add(1)
		go func(region string) {
			var info RegionSGs
			var err error
			defer wg.Done()
			account.Region = region
			sess, err := utils.GetSession(account)
			if err != nil {
				log.Println("Could not get security groups for", account.Profile, ":", err)
				return
			}
			info.SGs, err = GetRegionSGs(sess)
			if err != nil {
				log.Println("Could not get security groups for", region, "in", profile, ":", err)
				return
			}
			info.Region = region
			info.Profile = profile
			info.AccountID, err = utils.GetAccountId(sess)
			if err != nil {
				log.Println("Could not get account id for", account.Profile, ":", err)
				return
			}
			SGsChan <- info
		}(region)
	}

	go func() {
		wg.Wait()
		close(SGsChan)
	}()

	var accountSGs AccountSGs
	for regionSGs := range SGsChan {
		accountSGs = append(accountSGs, regionSGs)
	}

	return accountSGs, nil
}

func GetProfilesSGs(accounts []utils.AccountInfo) (ProfilesSGs, error) {
	profilesSGsChan := make(chan AccountSGs)
	var wg sync.WaitGroup

	for _, account := range accounts {
		wg.Add(1)
		go func(account utils.AccountInfo) {
			var err error
			defer wg.Done()
			accountSGs, err := GetAccountSGs(account)
			if err != nil {
				log.Println("Could not get security groups for", account.Profile, ":", err)
				return
			}
			profilesSGsChan <- accountSGs
		}(account)
	}

	go func() {
		wg.Wait()
		close(profilesSGsChan)
	}()

	var profilesSGs ProfilesSGs
	for accountSGs := range profilesSGsChan {
		profilesSGs = append(profilesSGs, accountSGs)
	}
	return profilesSGs, nil
}

//if a cidr is given, search the SGs for that rule and only print those containing the cidr
func WriteProfilesSGs(profileSGs ProfilesSGs, options SGOptions) error {
	outputDir := "output/ec2/"
	utils.MakeDir(outputDir)
	outputFile := outputDir + "sgs.csv"
	cidr := options.Cidr
	outfile, err := utils.CreateFile(outputFile)
	fmt.Println("Writing SGs to file:", outfile.Name())
	if err != nil {
		return fmt.Errorf("could not create sgs file", err)
	}

	writer := csv.NewWriter(outfile)
	defer writer.Flush()
	fmt.Println("Writing images to file:", outfile.Name())

	var columnTitles = []string{"Profile",
		"Account ID",
		"Region",
		"Security Group Name",
		"Security Group ID",
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

	for _, accountSGs := range profileSGs {
		for _, regionSGs := range accountSGs {
			for _, SG := range regionSGs.SGs {
				if cidr != "" {
					x := false
					for _, rules := range SG.IpPermissions {
						for _, ip := range rules.IpRanges {
							if *ip.CidrIp == cidr {
								x = true
							}
						}
					}
					if x {
						var data = []string{regionSGs.Profile,
							regionSGs.AccountID,
							regionSGs.Region,
							*SG.GroupName,
							*SG.GroupId,
						}

						if len(tags) > 0 {
							for _, tag := range tags {
								x := false
								for _, SGTag := range SG.Tags {
									if *SGTag.Key == tag {
										data = append(data, *SGTag.Value)
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
				} else {
					var data = []string{regionSGs.Profile,
						regionSGs.AccountID,
						regionSGs.Region,
						*SG.GroupName,
						*SG.GroupId,
					}

					if len(tags) > 0 {
						for _, tag := range tags {
							x := false
							for _, SGTag := range SG.Tags {
								if *SGTag.Key == tag {
									data = append(data, *SGTag.Value)
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
	}
	return nil
}
