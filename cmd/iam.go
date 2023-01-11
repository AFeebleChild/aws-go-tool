package cmd

import (
	"fmt"

	"github.com/afeeblechild/aws-go-tool/lib/iam"
	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/spf13/cobra"
)

var (
	Username  string
	RolesFile string
)

var iamCmd = &cobra.Command{
	Use:   "iam",
	Short: "For use with interacting with the iam service",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Run -h to see the help menu")
	},
}

var policiesListCmd = &cobra.Command{
	Use:   "policieslist",
	Short: "Will generate a report of policies",
	Run: func(cmd *cobra.Command, args []string) {
		profilesPolicies, err := iam.GetProfilesPolicies(Accounts)
		utils.Check(err)
		iam.WriteProfilesPolicies(profilesPolicies)
	},
}

var rolesListCmd = &cobra.Command{
	Use:   "roleslist",
	Short: "Will generate a report of roles",
	Run: func(cmd *cobra.Command, args []string) {
		profilesRoles, err := iam.GetProfilesRoles(Accounts)
		utils.Check(err)
		iam.WriteProfilesRoles(profilesRoles)
	},
}

var rolesUpdateCmd = &cobra.Command{
	Use:   "rolesupdate",
	Short: "Will update the roles session duration",
	Run: func(cmd *cobra.Command, args []string) {
		//TODO update this entire function
		//add cli parameter for duration
		err := iam.UpdateProfilesRolesSessionDuration(RolesFile, 28800)
		utils.Check(err)
	},
}

var usersListCmd = &cobra.Command{
	Use:   "userslist",
	Short: "Will generate a report of users",
	Run: func(cmd *cobra.Command, args []string) {
		profilesUsers, err := iam.GetProfilesUsers(Accounts)
		utils.Check(err)
		iam.WriteProfilesUsers(profilesUsers)
	},
}

//TODO reformat this func
var userUpdatePWCmd = &cobra.Command{
	Use:   "userupdatepw",
	Short: "Will update the users password",
	Run: func(cmd *cobra.Command, args []string) {
		for _, account := range Accounts {
			sess := utils.OpenSession(account.Profile, "us-east-1")
			user := iam.UserUpdate{Username: Username, ResetRequired: false}
			password, err := iam.UpdateUserPassword(user, sess)
			if err != nil {
				utils.LogAll("could not update pw:", err)
			}
			//TODO better output for passwords
			fmt.Println(password)
		}
	},
}

func init() {
	RootCmd.AddCommand(iamCmd)

	iamCmd.AddCommand(usersListCmd)
	iamCmd.AddCommand(userUpdatePWCmd)
	iamCmd.AddCommand(rolesListCmd)
	iamCmd.AddCommand(rolesUpdateCmd)
	iamCmd.AddCommand(policiesListCmd)

	RootCmd.PersistentFlags().StringVarP(&Username, "username", "u", "", "username to update")
	RootCmd.PersistentFlags().StringVarP(&RolesFile, "rolesfile", "f", "", "list of roles to update")
}
