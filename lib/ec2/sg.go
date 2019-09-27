package ec2

import (
	"encoding/csv"
	"fmt"
	"log"
	"strconv"
	"sync"

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
	Region    string
	Profile   string
	AccountID string
	SGs       []ec2.SecurityGroup
}

type AccountSGs []RegionSGs
type ProfilesSGs []AccountSGs

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
			sess, err := account.GetSession()
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

func WriteProfilesSgs(profileSGs ProfilesSGs, options SGOptions) error {
	outputDir := "output/ec2/"
	utils.MakeDir(outputDir)
	outputFile := outputDir + "sgs.csv"
	outfile, err := utils.CreateFile(outputFile)
	fmt.Println("Writing SGs to file:", outfile.Name())
	if err != nil {
		return fmt.Errorf("could not create sgs file", err)
	}

	writer := csv.NewWriter(outfile)
	defer writer.Flush()

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
	return nil
}

//if a cidr is given, search the SGs for that rule and only print those containing the cidr
func WriteProfilesSgRules(profileSGs ProfilesSGs, options SGOptions) error {
	outputDir := "output/ec2/"
	utils.MakeDir(outputDir)
	outputFile := outputDir + "sgRules.csv"
	cidr := options.Cidr
	outfile, err := utils.CreateFile(outputFile)
	fmt.Println("Writing SG Rules to file:", outfile.Name())
	if err != nil {
		return fmt.Errorf("could not create sgRules file", err)
	}

	writer := csv.NewWriter(outfile)
	defer writer.Flush()

	var columnTitles = []string{"Profile",
		"Account ID",
		"Region",
		"Security Group Name",
		"Security Group ID",
		"Rule Protocol",
		"Rule CIDR",
		"From Port",
		"To Port",
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
				//the cidr option has been passed so need to search for it
				if cidr != "" {
					for _, rule := range SG.IpPermissions {
						for _, ip := range rule.IpRanges {
							if *ip.CidrIp == cidr {
								var data = []string{regionSGs.Profile,
									regionSGs.AccountID,
									regionSGs.Region,
									*SG.GroupName,
									*SG.GroupId,
									*rule.IpProtocol,
									*ip.CidrIp,
									strconv.FormatInt(*rule.FromPort, 10),
									strconv.FormatInt(*rule.ToPort, 10),
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
					//no cidr option, just print all the rules
				} else {
					for _, rule := range SG.IpPermissions {
						var fromPort, toPort string
						if rule.FromPort != nil {
							fromPort = strconv.FormatInt(*rule.FromPort, 10)
						}
						if rule.ToPort != nil {
							toPort = strconv.FormatInt(*rule.ToPort, 10)
						}
						for _, ip := range rule.IpRanges {
							var data = []string{regionSGs.Profile,
								regionSGs.AccountID,
								regionSGs.Region,
								*SG.GroupName,
								*SG.GroupId,
								*rule.IpProtocol,
								*ip.CidrIp,
								fromPort,
								toPort,
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
		}
	}
	return nil
}
