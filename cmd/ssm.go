package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/afeeblechild/aws-go-tool/lib/ssm"
)

var ssmCmd = &cobra.Command{
	Use:   "ssm",
	Short: "For use with interacting with the ssm service",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Run -h to see the help menu")
	},
}

var patchCmd = &cobra.Command{
	Use:   "patch",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		accounts, err := utils.BuildAccountsSlice(ProfilesFile, AccessType)
		if err != nil {
			fmt.Println(err)
			return
		}

		OSList := []string{"WINDOWS", "SUSE", "CENTOS", "UBUNTU", "AMAZON_LINUX", "AMAZON_LINUX_2", "REDHAT_ENTERPRISE_LINUX"}
		ssm.AddRegionBaselines(OSList, accounts[0])
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		accounts, err := utils.BuildAccountsSlice(ProfilesFile, AccessType)
		if err != nil {
			fmt.Println(err)
			return
		}

		OSList := []string{"WINDOWS", "SUSE", "CENTOS", "UBUNTU", "AMAZON_LINUX", "AMAZON_LINUX_2", "REDHAT_ENTERPRISE_LINUX"}
		ssm.AddRegionBaselines(OSList, accounts[0])
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {
	RootCmd.AddCommand(ssmCmd)

	ssmCmd.AddCommand(patchCmd)
	ssmCmd.AddCommand(statusCmd)
}
