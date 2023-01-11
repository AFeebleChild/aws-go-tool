package cmd

import (
	"fmt"
	"os"

	"github.com/afeeblechild/aws-go-tool/lib/ssm"
	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/spf13/cobra"
)

var (
	AccountIdsFile string
	DocumentName   string
)

var ssmCmd = &cobra.Command{
	Use:   "ssm",
	Short: "For use with interacting with the ssm service",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Run -h to see the help menu")
	},
}

var removeDocumentPermissionsCmd = &cobra.Command{
	Use:   "removedocumentpermissions",
	Short: "remove permissions from private ssm document",
	Run: func(cmd *cobra.Command, args []string) {
		err := ssm.RemoveDocumentPermissionsFromAccounts(Accounts, AccountIdsFile, DocumentName)
		utils.Check(err)
	},
}

func init() {
	RootCmd.AddCommand(ssmCmd)

	ssmCmd.AddCommand(removeDocumentPermissionsCmd)

	ssmCmd.PersistentFlags().StringVarP(&AccountIdsFile, "accountidsfile", "f", "", "list of account ids to remove")
	ssmCmd.PersistentFlags().StringVarP(&DocumentName, "documentname", "d", "", "name of document to update")
}
