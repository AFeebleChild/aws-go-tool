package ec2

import (
	"encoding/csv"
	"fmt"
	"log"
	"sync"

	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type InstanceOptions struct {
	Tags []string
}

type RegionInstances struct {
	AccountId string
	Region    string
	Profile   string
	Instances []ec2.Instance
	Status    []ec2.InstanceStatus
}

type AccountInstances []RegionInstances
type ProfilesInstances []AccountInstances

// GetRegionInstances will take a session and pull all instances based on the region of the session
func (ri *RegionInstances) GetRegionInstances(sess *session.Session) error {
	svc := ec2.New(sess)
	params := &ec2.DescribeInstancesInput{}

	for {
		resp, err := svc.DescribeInstances(params)
		if err != nil {
			return err
		}

		for _, reservation := range resp.Reservations {
			for _, instance := range reservation.Instances {
				ri.Instances = append(ri.Instances, *instance)
			}
		}

		if resp.NextToken != nil {
			params.NextToken = resp.NextToken
		} else {
			break
		}
	}
	return nil
}

func (ri *RegionInstances) GetRegionInstancesStatuses(sess *session.Session) error {
	params := &ec2.DescribeInstanceStatusInput{}

	for {
		resp, err := ec2.New(sess).DescribeInstanceStatus(params)
		if err != nil {
			return err
		}

		for _, status := range resp.InstanceStatuses {
			ri.Status = append(ri.Status, *status)
		}

		if resp.NextToken != nil {
			params.NextToken = resp.NextToken
		} else {
			break
		}
	}

	return nil
}

// GetAccountInstances will take a profile and go through all regions to get all instances in the account
func GetAccountInstances(account utils.AccountInfo) (AccountInstances, error) {
	profile := account.Profile
	fmt.Println("Getting instances for profile:", profile)
	instancesChan := make(chan RegionInstances)
	var wg sync.WaitGroup

	for _, region := range utils.RegionMap {
		wg.Add(1)
		go func(region string) {
			defer wg.Done()
			info := &RegionInstances{AccountId: account.AccountId, Region: region, Profile: profile}
			sess, err := account.GetSession(region)
			if err != nil {
				log.Println("could not get session for", account.Profile, ":", err)
				return
			}
			if err = info.GetRegionInstances(sess); err != nil {
				log.Println("could not get instances for", region, "in", profile, ":", err)
				return
			}
			if err = info.GetRegionInstancesStatuses(sess); err != nil {
				log.Println("could not get instance statuses for", region, "in", profile, ":", err)
				return
			}
			instancesChan <- *info
		}(region)
	}

	go func() {
		wg.Wait()
		close(instancesChan)
	}()

	var accountInstances AccountInstances
	for regionInstances := range instancesChan {
		accountInstances = append(accountInstances, regionInstances)
	}

	return accountInstances, nil
}

// GetProfilesInstances will return all the instances in all accounts of a given filename with a list of profiles in it
func GetProfilesInstances(accounts []utils.AccountInfo) (ProfilesInstances, error) {
	profilesInstancesChan := make(chan AccountInstances)
	var wg sync.WaitGroup

	for _, account := range accounts {
		wg.Add(1)
		go func(account utils.AccountInfo) {
			defer wg.Done()
			if err := account.SetAccountId(); err != nil {
				log.Println("could not set account id for", account.Profile, ":", err)
				return
			}
			accountInstances, err := GetAccountInstances(account)
			if err != nil {
				log.Println("Could not get instances for", account.Profile, ":", err)
				return
			}
			profilesInstancesChan <- accountInstances
		}(account)
	}

	go func() {
		wg.Wait()
		close(profilesInstancesChan)
	}()

	var profilesInstances ProfilesInstances
	for accountInstances := range profilesInstancesChan {
		profilesInstances = append(profilesInstances, accountInstances)
	}
	return profilesInstances, nil
}

func WriteProfilesInstances(profileInstances ProfilesInstances, options utils.Ec2Options) error {
	outputDir := "output/ec2/"
	utils.MakeDir(outputDir)
	outputFile := outputDir + "instances.csv"
	outfile, err := utils.CreateFile(outputFile)
	if err != nil {
		return fmt.Errorf("could not create instances file: %v", err)
	}

	fmt.Println("Writing instances to file:", outfile.Name())
	writer := csv.NewWriter(outfile)
	defer writer.Flush()
	var columnTitles = []string{"Profile",
		"Account ID",
		"Region",
		"Instance Name",
		"Instance ID",
		"Private IP",
		"Public IP",
		"Pem Key",
		"Instance Type",
		"Instance State",
		"AMI ID",
		"VPC",
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

	for _, accountInstances := range profileInstances {
		for _, regionInstances := range accountInstances {
			for _, instance := range regionInstances.Instances {
				var instanceName string
				for _, tag := range instance.Tags {
					if *tag.Key == "Name" {
						instanceName = *tag.Value
					}
				}
				privateIp := ""
				if instance.PrivateIpAddress != nil {
					privateIp = *instance.PrivateIpAddress
				}
				pemKey := ""
				if instance.KeyName != nil {
					pemKey = *instance.KeyName
				}

				var vpcId string
				if instance.VpcId != nil {
					vpcId = *instance.VpcId
				}

				publicIp := "N/A"
				if instance.PublicIpAddress != nil {
					publicIp = *instance.PublicIpAddress
				}

				amiId := "N/A"
				if instance.ImageId != nil {
					amiId = *instance.ImageId
				}

				var data = []string{regionInstances.Profile,
					regionInstances.AccountId,
					regionInstances.Region,
					instanceName,
					*instance.InstanceId,
					privateIp,
					publicIp,
					pemKey,
					*instance.InstanceType,
					*instance.State.Name,
					amiId,
					vpcId,
				}

				if len(tags) > 0 {
					for _, tag := range tags {
						x := false
						for _, instanceTag := range instance.Tags {
							if *instanceTag.Key == tag {
								data = append(data, *instanceTag.Value)
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
