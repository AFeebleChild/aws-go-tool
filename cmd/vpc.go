// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/afeeblechild/aws-go-tool/lib/vpc"

	"github.com/spf13/cobra"
)

// vpcCmd represents the vpc command
var vpcCmd = &cobra.Command{
	Use:   "vpc",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		fmt.Println("vpc called")
	},
}

var subnetsListCmd = &cobra.Command{
	Use:   "subnetslist",
	Short: "Will generate a report of vpc info for all given accounts",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		accounts, err := utils.BuildAccountsSlice(ProfilesFile, AccessType)
		if err != nil {
			fmt.Println(err)
			return
		}

		profilesSubnets, err := vpc.GetProfilesSubnets(accounts)
		if err != nil {
			fmt.Println(err)
			return
		}
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
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

var vpcsListCmd = &cobra.Command{
	Use:   "vpcslist",
	Short: "Will generate a report of vpc info for all given accounts",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		accounts, err := utils.BuildAccountsSlice(ProfilesFile, AccessType)
		if err != nil {
			fmt.Println(err)
			return
		}

		profilesVpcs, err := vpc.GetProfilesVpcs(accounts)
		if err != nil {
			fmt.Println(err)
			return
		}
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// vpcCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// vpcCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
