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
			sess, err := utils.GetSession(account)
			if err != nil {
				log.Println("Could not get session for", account.Profile, ":", err)
				return
			}
			info.Instances, err = GetRegionInstances(sess)
			info.Status, err = GetRegionInstancesStatuses(sess)
			if err != nil {
				log.Println("Could not get instances for", region, "in", profile, ":", err)
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
	outfile, err := utils.CreateFile("instances.csv")
	if err != nil {
		return fmt.Errorf("could not create instances file", err)
	}

	writer := csv.NewWriter(outfile)
	defer writer.Flush()
	fmt.Println("Writing instances to file:", outfile.Name())
	var columnTitles = []string{"Profile",
	//TODO add account ID
		"Region",
		"Instance Name",
		"Instance ID",
		"Private IP",
		"Pem Key",
		"Instance Type",
		"Instance State",
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
				privateIP := ""
				if instance.PrivateIpAddress != nil {
					privateIP = *instance.PrivateIpAddress
				}
				pemKey := ""
				if instance.KeyName != nil {
					pemKey = *instance.KeyName
				}

				//x := true
				//for _, key := range pemKeys {
				//	if pemKey == key {
				//		x = false
				//		break
				//	}
				//}
				//if x {
				//	xx := true
				//	for _, tag := range instance.Tags {
				//		if *tag.Key == "infra_msp" && *tag.Value != "2w"{
				//			xx = false
				//		}
				//	}
				//	if xx {
				//		pemKeys = append(pemKeys, pemKey)
				//		fmt.Fprintln(pemKeyFile, pemKey, ",", regionInstances.Profile)
				//	}
				//}

				//If the instance is running, get the instance state
				//Code 16 is instance running
				//var systemCheck, instanceCheck string
				//if *instance.State.Code == 16 {
				//	for _, status := range regionInstances.Status {
				//		if *instance.InstanceId == *status.InstanceId {
				//			systemCheck = *status.SystemStatus.Status
				//			instanceCheck = *status.InstanceStatus.Status
				//		}
				//	}
				//}

				var vpcID string
				if instance.VpcId != nil {
					vpcID = *instance.VpcId
				}

				var data = []string{regionInstances.Profile,
					regionInstances.Region,
					instanceName,
					*instance.InstanceId,
					privateIP,
					pemKey,
					*instance.InstanceType,
					*instance.State.Name,
					vpcID,
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
