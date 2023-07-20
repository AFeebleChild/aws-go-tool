package cmd

import (
	"fmt"

	"github.com/afeeblechild/aws-go-tool/lib/ec2"
	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/spf13/cobra"
)

var (
	Cidr string
)

var ec2Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "For use with interacting with the ec2 service",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Run -h to see the help menu")
	},
}

var imagesCheckCmd = &cobra.Command{
	Use:   "imagescheck",
	Short: "Will generate a report of images in use by instances in the account.",
	Run: func(cmd *cobra.Command, args []string) {
		checkedImages, err := ec2.CheckImages(Accounts)
		if err != nil {
			fmt.Println(err)
			return
		}

		options := utils.Ec2Options{Tags: Tags}
		err = ec2.WriteCheckedImages(checkedImages, options)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

var imagesListCmd = &cobra.Command{
	Use:   "imageslist",
	Short: "Will generate a report of all images for all given accounts.",
	Run: func(cmd *cobra.Command, args []string) {
		profilesImages, err := ec2.GetProfilesImages(Accounts)
		if err != nil {
			fmt.Println(err)
			return
		}

		options := utils.Ec2Options{Tags: Tags}
		err = ec2.WriteProfilesImages(profilesImages, options)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

var instancesListCmd = &cobra.Command{
	Use:   "instanceslist",
	Short: "Will generate a report of all instances for all given accounts.",
	Run: func(cmd *cobra.Command, args []string) {
		profilesInstances, err := ec2.GetProfilesInstances(Accounts)
		if err != nil {
			fmt.Println(err)
			return
		}
		options := utils.Ec2Options{Tags: Tags}
		err = ec2.WriteProfilesInstances(profilesInstances, options)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

var sgsListCmd = &cobra.Command{
	Use:   "sgslist",
	Short: "Will generate a report of all security groups for all given accounts.",
	Run: func(cmd *cobra.Command, args []string) {
		profilesSGs, err := ec2.GetProfilesSGs(Accounts)
		if err != nil {
			fmt.Println(err)
			return
		}
		options := ec2.SgOptions{Tags: Tags}
		err = ec2.WriteProfilesSgs(profilesSGs, options)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

var sgsRulesListCmd = &cobra.Command{
	Use:   "sgruleslist",
	Short: "Will generate a report of all security group rules for all given accounts",
	Run: func(cmd *cobra.Command, args []string) {
		profilesSGs, err := ec2.GetProfilesSGs(Accounts)
		if err != nil {
			fmt.Println(err)
			return
		}
		options := ec2.SgOptions{Tags: Tags}
		options.Cidr = Cidr
		err = ec2.WriteProfilesSgRules(profilesSGs, options)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

var snapshotsListCmd = &cobra.Command{
	Use:   "snapshotslist",
	Short: "Will generate a report of all snapshots for all given accounts.",
	Run: func(cmd *cobra.Command, args []string) {
		profilesSnapshots, err := ec2.GetProfilesSnapshots(Accounts)
		if err != nil {
			fmt.Println(err)
			return
		}
		options := utils.Ec2Options{Tags: Tags}
		err = ec2.WriteProfilesSnapshots(profilesSnapshots, options)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

var volumesListCmd = &cobra.Command{
	Use:   "volumeslist",
	Short: "Will generate a report of all volumes for all given accounts.",
	Run: func(cmd *cobra.Command, args []string) {
		profilesVolumes, err := ec2.GetProfilesVolumes(Accounts)
		if err != nil {
			fmt.Println(err)
			return
		}
		options := utils.Ec2Options{Tags: Tags}
		err = ec2.WriteProfilesVolumes(profilesVolumes, options)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {
	RootCmd.AddCommand(ec2Cmd)

	ec2Cmd.AddCommand(imagesCheckCmd)
	ec2Cmd.AddCommand(imagesListCmd)
	ec2Cmd.AddCommand(instancesListCmd)
	ec2Cmd.AddCommand(sgsListCmd)
	ec2Cmd.AddCommand(sgsRulesListCmd)
	ec2Cmd.AddCommand(snapshotsListCmd)
	ec2Cmd.AddCommand(volumesListCmd)

	sgsRulesListCmd.PersistentFlags().StringVarP(&Cidr, "cidr", "c", "", "cidr to search for")
}
