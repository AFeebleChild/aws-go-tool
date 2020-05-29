package cmd

import (
	"fmt"
	"log"

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
		accounts, err := utils.BuildAccountsSlice(ProfilesFile, AccessType)
		if err != nil {
			fmt.Println(err)
			return
		}

		checkedImages, err := ec2.CheckImages(accounts)

		var tags []string
		if TagFile != "" {
			tags, err = utils.ReadFile(TagFile)
			if err != nil {
				log.Println("could not open tagFile:", err, "\ncontinuing without tags in output")
				fmt.Println("could not open tagFile:", err, "\ncontinuing without tags in output")
			}
		}
		options := utils.Ec2Options{Tags: tags}
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
		accounts, err := utils.BuildAccountsSlice(ProfilesFile, AccessType)
		if err != nil {
			fmt.Println(err)
			return
		}

		profilesImages, err := ec2.GetProfilesImages(accounts)
		if err != nil {
			fmt.Println(err)
			return
		}
		var tags []string
		if TagFile != "" {
			tags, err = utils.ReadFile(TagFile)
			if err != nil {
				log.Println("could not open tagFile:", err, "\ncontinuing without tags in output")
				fmt.Println("could not open tagFile:", err, "\ncontinuing without tags in output")
			}
		}
		options := utils.Ec2Options{Tags: tags}
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
		accounts, err := utils.BuildAccountsSlice(ProfilesFile, AccessType)
		if err != nil {
			fmt.Println(err)
			return
		}

		profilesInstances, err := ec2.GetProfilesInstances(accounts)
		if err != nil {
			fmt.Println(err)
			return
		}
		var tags []string
		if TagFile != "" {
			tags, err = utils.ReadFile(TagFile)
			if err != nil {
				log.Println("could not open tagFile:", err, "\ncontinuing without tags in output")
				fmt.Println("could not open tagFile:", err, "\ncontinuing without tags in output")
			}
		}
		options := utils.Ec2Options{Tags: tags}
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
		accounts, err := utils.BuildAccountsSlice(ProfilesFile, AccessType)
		if err != nil {
			fmt.Println(err)
			return
		}

		profilesSGs, err := ec2.GetProfilesSGs(accounts)
		if err != nil {
			fmt.Println(err)
			return
		}
		var tags []string
		if TagFile != "" {
			tags, err = utils.ReadFile(TagFile)
			if err != nil {
				log.Println("could not open tagFile:", err, "\ncontinuing without tags in output")
				fmt.Println("could not open tagFile:", err, "\ncontinuing without tags in output")
			}
		}
		options := ec2.SGOptions{Tags: tags}
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
		accounts, err := utils.BuildAccountsSlice(ProfilesFile, AccessType)
		if err != nil {
			fmt.Println(err)
			return
		}

		profilesSGs, err := ec2.GetProfilesSGs(accounts)
		if err != nil {
			fmt.Println(err)
			return
		}
		var tags []string
		if TagFile != "" {
			tags, err = utils.ReadFile(TagFile)
			if err != nil {
				log.Println("could not open tagFile:", err, "\ncontinuing without tags in output")
				fmt.Println("could not open tagFile:", err, "\ncontinuing without tags in output")
			}
		}
		options := ec2.SGOptions{Tags: tags}
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
		accounts, err := utils.BuildAccountsSlice(ProfilesFile, AccessType)
		if err != nil {
			fmt.Println(err)
			return
		}

		profilesSnapshots, err := ec2.GetProfilesSnapshots(accounts)
		if err != nil {
			fmt.Println(err)
			return
		}
		var tags []string
		if TagFile != "" {
			tags, err = utils.ReadFile(TagFile)
			if err != nil {
				log.Println("could not open tagFile:", err, "\ncontinuing without tags in output")
				fmt.Println("could not open tagFile:", err, "\ncontinuing without tags in output")
			}
		}
		options := utils.Ec2Options{Tags: tags}
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
		accounts, err := utils.BuildAccountsSlice(ProfilesFile, AccessType)
		if err != nil {
			fmt.Println(err)
			return
		}

		profilesVolumes, err := ec2.GetProfilesVolumes(accounts)
		if err != nil {
			fmt.Println(err)
			return
		}
		var tags []string
		if TagFile != "" {
			tags, err = utils.ReadFile(TagFile)
			if err != nil {
				log.Println("could not open tagFile:", err, "\ncontinuing without tags in output")
				fmt.Println("could not open tagFile:", err, "\ncontinuing without tags in output")
			}
		}
		options := utils.Ec2Options{Tags: tags}
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
