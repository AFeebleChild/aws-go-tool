package ec2

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

type SgOptions struct {
	Cidr string
	Tags []string
}

type RegionSecurityGroups struct {
	Region         string
	Profile        string
	AccountId      string
	SecurityGroups []ec2.SecurityGroup
}

type AccountSecurityGroups []RegionSecurityGroups
type ProfilesSecurityGroups []AccountSecurityGroups

func (rs *RegionSecurityGroups) GetRegionSecurityGroups(sess *session.Session) error {
	svc := ec2.New(sess)
	params := &ec2.DescribeSecurityGroupsInput{}

	for {
		resp, err := svc.DescribeSecurityGroups(params)
		if err != nil {
			return err
		}

		for _, SG := range resp.SecurityGroups {
			rs.SecurityGroups = append(rs.SecurityGroups, *SG)
		}

		if resp.NextToken != nil {
			params.NextToken = resp.NextToken
		} else {
			break
		}
	}

	return nil
}

func GetAccountSecurityGroups(account utils.AccountInfo) (AccountSecurityGroups, error) {
	profile := account.Profile
	fmt.Println("Getting Security Groups for profile:", profile)
	SGsChan := make(chan RegionSecurityGroups)
	var wg sync.WaitGroup

	for _, region := range utils.RegionMap {
		wg.Add(1)
		go func(region string) {
			defer wg.Done()
			info := RegionSecurityGroups{AccountId: account.AccountId, Region: region, Profile: profile}
			sess, err := account.GetSession(region)
			if err != nil {
				log.Println("Could not get security groups for", account.Profile, ":", err)
				return
			}
			if err = info.GetRegionSecurityGroups(sess); err != nil {
				log.Println("Could not get security groups for", region, "in", profile, ":", err)
				return
			}
			SGsChan <- info
		}(region)
	}

	go func() {
		wg.Wait()
		close(SGsChan)
	}()

	var accountSGs AccountSecurityGroups
	for regionSGs := range SGsChan {
		accountSGs = append(accountSGs, regionSGs)
	}

	return accountSGs, nil
}

func GetProfilesSGs(accounts []utils.AccountInfo) (ProfilesSecurityGroups, error) {
	profilesSGsChan := make(chan AccountSecurityGroups)
	var wg sync.WaitGroup

	for _, account := range accounts {
		wg.Add(1)
		go func(account utils.AccountInfo) {
			defer wg.Done()
			if err := account.SetAccountId(); err != nil {
				log.Println("could not set account id for", account.Profile, ":", err)
				return
			}
			accountSGs, err := GetAccountSecurityGroups(account)
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

	var profilesSGs ProfilesSecurityGroups
	for accountSGs := range profilesSGsChan {
		profilesSGs = append(profilesSGs, accountSGs)
	}
	return profilesSGs, nil
}

func WriteProfilesSgs(profileSGs ProfilesSecurityGroups, options SgOptions) error {
	outputDir := "output/ec2/"
	utils.MakeDir(outputDir)
	outputFile := outputDir + "sgs.csv"
	outfile, err := utils.CreateFile(outputFile)
	if err != nil {
		return fmt.Errorf("could not create sgs file: %v", err)
	}

	fmt.Println("Writing SGs to file:", outfile.Name())
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

	if err = writer.Write(columnTitles); err != nil {
		fmt.Println(err)
	}

	for _, accountSGs := range profileSGs {
		for _, regionSGs := range accountSGs {
			for _, SG := range regionSGs.SecurityGroups {
				var data = []string{regionSGs.Profile,
					regionSGs.AccountId,
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

				if err = writer.Write(data); err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	return nil
}

// if a cidr is given, search the SGs for that rule and only print those containing the cidr
func WriteProfilesSgRules(profileSGs ProfilesSecurityGroups, options SgOptions) error {
	outputDir := "output/ec2/"
	utils.MakeDir(outputDir)
	outputFile := outputDir + "sgRules.csv"
	cidr := options.Cidr
	outfile, err := utils.CreateFile(outputFile)
	if err != nil {
		return fmt.Errorf("could not create sgRules file: %v", err)
	}

	fmt.Println("Writing SG Rules to file:", outfile.Name())
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

	if err = writer.Write(columnTitles); err != nil {
		fmt.Println(err)
	}

	for _, accountSGs := range profileSGs {
		for _, regionSGs := range accountSGs {
			for _, SG := range regionSGs.SecurityGroups {
				//the cidr option has been passed so need to search for it
				if cidr != "" {
					for _, rule := range SG.IpPermissions {
						for _, ip := range rule.IpRanges {
							if *ip.CidrIp == cidr {
								var data = []string{regionSGs.Profile,
									regionSGs.AccountId,
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
								regionSGs.AccountId,
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

							if err = writer.Write(data); err != nil {
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
