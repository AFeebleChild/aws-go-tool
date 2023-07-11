package cmd

import (
	"fmt"

	"github.com/afeeblechild/aws-go-tool/lib/vpc"
	"github.com/spf13/cobra"
)

var vpcCmd = &cobra.Command{
	Use:   "vpc",
	Short: "For use with interacting with the vpc service",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Run -h to see the help menu")
	},
}

var subnetsListCmd = &cobra.Command{
	Use:   "subnetslist",
	Short: "Will generate a report of vpc info for all given accounts",
	Run: func(cmd *cobra.Command, args []string) {
		profilesSubnets, err := vpc.GetProfilesSubnets(Accounts)
		if err != nil {
			fmt.Println(err)
			return
		}
		//options := utils.Ec2Options{Tags:Tags}
		err = vpc.WriteProfilesSubnets(profilesSubnets)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

var vpcsListCmd = &cobra.Command{
	Use:   "vpcslist",
	Short: "Will generate a report of vpc info for all given accounts",
	Run: func(cmd *cobra.Command, args []string) {
		profilesVpcs, err := vpc.GetProfilesVpcs(Accounts)
		if err != nil {
			fmt.Println(err)
			return
		}
		//options := utils.Ec2Options{Tags:Tags}
		err = vpc.WriteProfilesVpcs(profilesVpcs)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {
	RootCmd.AddCommand(vpcCmd)

	vpcCmd.AddCommand(vpcsListCmd)
	vpcCmd.AddCommand(subnetsListCmd)
}
