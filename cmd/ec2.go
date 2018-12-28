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

	"github.com/afeeblechild/aws-go-tool/lib/ec2"
	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/spf13/cobra"
	"log"
)

var (
	Cidr string
)

// ec2Cmd represents the ec2 command
var ec2Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "For use with interacting with the ec2 service",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Run -h to see the help menu")
	},
}

var snapshotsListCmd = &cobra.Command{
	Use:   "snapshotslist",
	Short: "Will generate a report of all snapshots for all given accounts",
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
				fmt.Println("could not open tagFile:", err)
				fmt.Println("continuing without tags in output")
			}
		}
		options := utils.Ec2Options{Tags:tags}
		err = ec2.WriteProfilesSnapshots(profilesSnapshots, options)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

var volumesListCmd = &cobra.Command{
	Use:   "volumeslist",
	Short: "Will generate a report of all volumes for all given accounts",
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
				fmt.Println("could not open tagFile:", err)
				fmt.Println("continuing without tags in output")
			}
		}
		options := utils.Ec2Options{Tags:tags}
		err = ec2.WriteProfilesVolumes(profilesVolumes, options)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

var imagesListCmd = &cobra.Command{
	Use:   "imageslist",
	Short: "Will generate a report of all amis for all given accounts",
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

		profilesImages, err := ec2.GetProfilesImages(accounts)
		if err != nil {
			fmt.Println(err)
			return
		}
		var tags []string
		if TagFile != "" {
			tags, err = utils.ReadFile(TagFile)
			if err != nil {
				log.Println("could not open tagFile:", err, "\ncontinuuing without tags in output")
				fmt.Println("could not open tagFile:", err)
				fmt.Println("continuing without tags in output")
			}
		}
		options := utils.Ec2Options{Tags:tags}
		err = ec2.WriteProfilesImages(profilesImages, options)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

var instancesListCmd = &cobra.Command{
	Use:   "instanceslist",
	Short: "Will generate a report of all instances for all given accounts",
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
				fmt.Println("could not open tagFile:", err)
				fmt.Println("continuing without tags in output")
			}
		}
		options := utils.Ec2Options{Tags:tags}
		err = ec2.WriteProfilesInstances(profilesInstances, options)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

var sgsListCmd = &cobra.Command{
	Use:   "sgslist",
	Short: "Will generate a report of all security groups for all given accounts",
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

		profilesSGs, err := ec2.GetProfilesSGs(accounts)
		if err != nil {
			fmt.Println(err)
			return
		}
		var tags []string
		if TagFile != "" {
			tags, err = utils.ReadFile(TagFile)
			if err != nil {
				log.Println("could not open tagFile:", err, "\ncontinuuing without tags in output")
				fmt.Println("could not open tagFile:", err)
				fmt.Println("continuing without tags in output")
			}
		}
		options := ec2.SGOptions{Tags:tags}
		err = ec2.WriteProfilesSgs(profilesSGs, options)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

var sgRulesListCmd = &cobra.Command{
	Use:   "sgruleslist",
	Short: "Will generate a report of all security group rules for all given accounts",
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

		profilesSGs, err := ec2.GetProfilesSGs(accounts)
		if err != nil {
			fmt.Println(err)
			return
		}
		var tags []string
		if TagFile != "" {
			tags, err = utils.ReadFile(TagFile)
			if err != nil {
				log.Println("could not open tagFile:", err, "\ncontinuuing without tags in output")
				fmt.Println("could not open tagFile:", err)
				fmt.Println("continuing without tags in output")
			}
		}
		options := ec2.SGOptions{Tags:tags}
		options.Cidr = Cidr
		err = ec2.WriteProfilesSgRules(profilesSGs, options)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {
	RootCmd.AddCommand(ec2Cmd)

	ec2Cmd.AddCommand(snapshotsListCmd)
	ec2Cmd.AddCommand(volumesListCmd)
	ec2Cmd.AddCommand(imagesListCmd)
	ec2Cmd.AddCommand(instancesListCmd)
	ec2Cmd.AddCommand(sgsListCmd)
	ec2Cmd.AddCommand(sgRulesListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ec2Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ec2Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	sgRulesListCmd.PersistentFlags().StringVarP(&Cidr, "cidr", "c", "", "cidr to search for")
}
