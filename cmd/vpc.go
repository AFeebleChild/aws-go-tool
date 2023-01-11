package cmd

import (
	"fmt"

	"github.com/afeeblechild/aws-go-tool/lib/utils"
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
		utils.Check(err)
		//var tags []string
		//if TagFile != "" {
		//	tags, err = utils.ReadFile(TagFile)
		//	if err != nil {
		//		log.Println("could not open tagFile:", err, "\ncontinuing without tags in output")
		//		fmt.Println("could not open tagFile:", err)
		//		fmt.Println("continuing without tags in output")
		//	}
		//}
		//options := utils.Ec2Options{Tags:tags}
		err = vpc.WriteProfilesSubnets(profilesSubnets)
		utils.Check(err)
	},
}

var vpcsListCmd = &cobra.Command{
	Use:   "vpcslist",
	Short: "Will generate a report of vpc info for all given accounts",
	Run: func(cmd *cobra.Command, args []string) {
		profilesVpcs, err := vpc.GetProfilesVpcs(Accounts)
		utils.Check(err)
		//var tags []string
		//if TagFile != "" {
		//	tags, err = utils.ReadFile(TagFile)
		//	if err != nil {
		//		log.Println("could not open tagFile:", err, "\ncontinuing without tags in output")
		//		fmt.Println("could not open tagFile:", err)
		//		fmt.Println("continuing without tags in output")
		//	}
		//}
		//options := utils.Ec2Options{Tags:tags}
		err = vpc.WriteProfilesVpcs(profilesVpcs)
		utils.Check(err)
	},
}

func init() {
	RootCmd.AddCommand(vpcCmd)

	vpcCmd.AddCommand(vpcsListCmd)
	vpcCmd.AddCommand(subnetsListCmd)
}
