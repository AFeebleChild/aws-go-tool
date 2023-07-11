package ec2

import (
	"encoding/csv"
	"fmt"
	"log"
	"sync"

	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/aws/aws-sdk-go/aws"
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

//GetRegionInstances will take a session and pull all instances based on the region of the session
func GetRegionInstances(sess *session.Session) ([]ec2.Instance, error) {
	var instances []ec2.Instance
	params := &ec2.DescribeInstancesInput{
		DryRun: aws.Bool(false),
	}

	resp, err := ec2.New(sess).DescribeInstances(params)
	if err != nil {
		return nil, err
	}

	//Extract the instances from the reservations
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			instances = append(instances, *instance)
		}
	}

	return instances, nil
}

func GetRegionInstancesStatuses(sess *session.Session) ([]ec2.InstanceStatus, error) {
	var instanceStatuses []ec2.InstanceStatus
	params := &ec2.DescribeInstanceStatusInput{
		DryRun: aws.Bool(false),
	}

	resp, err := ec2.New(sess).DescribeInstanceStatus(params)
	if err != nil {
		return nil, err
	}

	for _, status := range resp.InstanceStatuses {
		instanceStatuses = append(instanceStatuses, *status)
	}

	return instanceStatuses, nil
}

//GetAccountInstances will take a profile and go through all regions to get all instances in the account
func GetAccountInstances(account utils.AccountInfo) (AccountInstances, error) {
	profile := account.Profile
	fmt.Println("Getting instances for profile:", profile)
	instancesChan := make(chan RegionInstances)
	var wg sync.WaitGroup

	for _, region := range utils.RegionMap {
		wg.Add(1)
		go func(region string) {
			var info RegionInstances
			var err error
			defer wg.Done()
			account.Region = region
			sess, err := account.GetSession()
			if err != nil {
				log.Println("could not get session for", account.Profile, ":", err)
				return
			}
			info.Instances, err = GetRegionInstances(sess)
			info.Status, err = GetRegionInstancesStatuses(sess)
			if err != nil {
				log.Println("could not get instances for", region, "in", profile, ":", err)
				return
			}
			info.AccountId, err = utils.GetAccountId(sess)
			if err != nil {
				log.Println("could not get account id for profile: ", profile)
				return
			}
			info.Region = region
			info.Profile = profile
			instancesChan <- info
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

//GetProfilesInstances will return all the instances in all accounts of a given filename with a list of profiles in it
func GetProfilesInstances(accounts []utils.AccountInfo) (ProfilesInstances, error) {
	profilesInstancesChan := make(chan AccountInstances)
	var wg sync.WaitGroup

	for _, account := range accounts {
		wg.Add(1)
		go func(account utils.AccountInfo) {
			var err error
			defer wg.Done()
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

	writer := csv.NewWriter(outfile)
	defer writer.Flush()
	fmt.Println("Writing instances to file:", outfile.Name())
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

	err = writer.Write(columnTitles)
	if err != nil {
		fmt.Println(err)
	}
	//var pemKeys []string
	//pemKeyFile, _ := utils.CreateFile("pemKeys.csv")
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

				err = writer.Write(data)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	return nil
}
