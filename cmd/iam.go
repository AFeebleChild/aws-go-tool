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

	"github.com/afeeblechild/aws-go-tool/lib/iam"
	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/spf13/cobra"
)

// iamCmd represents the iam command
var iamCmd = &cobra.Command{
	Use:   "iam",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		fmt.Println("iam called")
	},
}

var usersListCmd = &cobra.Command{
	Use:   "userslist",
	Short: "A brief description of your command",
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

		profilesUsers, err := iam.GetProfilesUsers(accounts)
		if err != nil {
			fmt.Println("Could not get users from all profiles", err)
			return
		}
		iam.WriteProfilesUsers(profilesUsers)
	},
}

var userUpdatePWCmd = &cobra.Command{
	Use:   "userupdatepw",
	Short: "A brief description of your command",
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

		for _, account := range accounts {
			sess := utils.OpenSession(account.Profile, "us-east-1")
			user := iam.UserUpdate{Username: "atlas@2ndwatch", ResetRequired: false}
			password, err := iam.UpdateUserPassword(user, sess)
			if err != nil {
				//TODO better logging
				fmt.Println("could not update pw:", err)
			}
			//TODO better output for passwords
			fmt.Println(password)
		}
	},
}

var rolesListCmd = &cobra.Command{
	Use:   "roleslist",
	Short: "A brief description of your command",
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

		profilesRoles, err := iam.GetProfilesRoles(accounts)
		if err != nil {
			fmt.Println("Could not get users from all profiles", err)
			return
		}
		iam.WriteProfilesRoles(profilesRoles)
	},
}

func init() {
	RootCmd.AddCommand(iamCmd)

	iamCmd.AddCommand(usersListCmd)
	iamCmd.AddCommand(userUpdatePWCmd)
	iamCmd.AddCommand(rolesListCmd)

	//usersCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// iamCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// iamCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
