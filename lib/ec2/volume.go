package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws/session"
)

type RegionVolumes struct {
	Region    string
	Profile   string
	Volumes []ec2.Volume
}

type AccountVolumes []RegionVolumes
type ProfilesVolumes []AccountVolumes

//GetRegionVolumes will take a session and get all volumes based on the region of the session
func GetRegionVolumes(sess *session.Session) ([]ec2.Volume, error) {
	var volumes []ec2.Volume
	params := &ec2.DescribeVolumesInput{
		DryRun:   aws.Bool(false),
	}

	resp, err := ec2.New(sess).DescribeVolumes(params)
	if err != nil {
		return nil, err
	}

	//Add the volumes from the response to a slice to return
	for _, volume := range resp.Volumes {
		volumes = append(volumes, *volume)
	}

	return volumes, nil
}
