package cmd

import (
	"fmt"

	"github.com/afeeblechild/aws-go-tool/lib/ec2"
	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/spf13/cobra"
)

var (
	Cidr    string
	ec2Tags []string
)

var ec2Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "For use with interacting with the ec2 service",
	PersistentPreRun: func(cmd *cobra.Command, args []string){
		var err error
		if TagFile != "" {
			ec2Tags, err = utils.ReadFile(TagFile)
			if err != nil {
				utils.LogAll("could not open tagFile: ", err, "\ncontinuing without tags in output")
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Run -h to see the help menu")
	},
}

var imagesCheckCmd = &cobra.Command{
	Use:   "imagescheck",
	Short: "Will generate a report of images in use by instances in the account.",
	Run: func(cmd *cobra.Command, args []string) {
		checkedImages, err := ec2.CheckImages(Accounts)
		utils.Check(err)

		options := utils.Ec2Options{Tags: ec2Tags}
		err = ec2.WriteCheckedImages(checkedImages, options)
		utils.Check(err)
	},
}

var imagesListCmd = &cobra.Command{
	Use:   "imageslist",
	Short: "Will generate a report of all images for all given accounts.",
	Run: func(cmd *cobra.Command, args []string) {
		profilesImages, err := ec2.GetProfilesImages(Accounts)
		utils.Check(err)

		options := utils.Ec2Options{Tags: ec2Tags}
		err = ec2.WriteProfilesImages(profilesImages, options)
		utils.Check(err)
	},
}

var instancesListCmd = &cobra.Command{
	Use:   "instanceslist",
	Short: "Will generate a report of all instances for all given accounts.",
	Run: func(cmd *cobra.Command, args []string) {
		profilesInstances, err := ec2.GetProfilesInstances(Accounts)
		utils.Check(err)

		options := utils.Ec2Options{Tags: ec2Tags}
		err = ec2.WriteProfilesInstances(profilesInstances, options)
		utils.Check(err)
	},
}

var sgsListCmd = &cobra.Command{
	Use:   "sgslist",
	Short: "Will generate a report of all security groups for all given accounts.",
	Run: func(cmd *cobra.Command, args []string) {
		profilesSGs, err := ec2.GetProfilesSGs(Accounts)
		utils.Check(err)

		options := ec2.SGOptions{Tags: ec2Tags}
		err = ec2.WriteProfilesSgs(profilesSGs, options)
		utils.Check(err)
	},
}

var sgsRulesListCmd = &cobra.Command{
	Use:   "sgruleslist",
	Short: "Will generate a report of all security group rules for all given accounts",
	Run: func(cmd *cobra.Command, args []string) {
		profilesSGs, err := ec2.GetProfilesSGs(Accounts)
		utils.Check(err)

		options := ec2.SGOptions{Tags: ec2Tags}
		options.Cidr = Cidr
		err = ec2.WriteProfilesSgRules(profilesSGs, options)
		utils.Check(err)
	},
}

var snapshotsListCmd = &cobra.Command{
	Use:   "snapshotslist",
	Short: "Will generate a report of all snapshots for all given accounts.",
	Run: func(cmd *cobra.Command, args []string) {
		profilesSnapshots, err := ec2.GetProfilesSnapshots(Accounts)
		utils.Check(err)

		options := utils.Ec2Options{Tags: ec2Tags}
		err = ec2.WriteProfilesSnapshots(profilesSnapshots, options)
		utils.Check(err)
	},
}

var volumesListCmd = &cobra.Command{
	Use:   "volumeslist",
	Short: "Will generate a report of all volumes for all given accounts.",
	Run: func(cmd *cobra.Command, args []string) {
		profilesVolumes, err := ec2.GetProfilesVolumes(Accounts)
		utils.Check(err)

		options := utils.Ec2Options{Tags: ec2Tags}
		err = ec2.WriteProfilesVolumes(profilesVolumes, options)
		utils.Check(err)
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
