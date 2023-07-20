package ssm

import (
	"fmt"

	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

//Requires the list of accounts ids in a file, with 1 account id per line
func RemoveDocumentPermissionsFromAccounts(accounts []utils.AccountInfo, accountIdsFile string, documentName string) error {
	accountIds, err := utils.ReadFile(accountIdsFile)
	if err != nil {
		return fmt.Errorf("could not read accountIdsFile: %v", err)
	}

	var pointerAccountIds []*string
	for _, accountId := range accountIds {
		pointerAccountIds = append(pointerAccountIds, aws.String(accountId))
	}

	for _, account := range accounts {
		sess, err := account.GetSession("us-east-1")
		if err != nil {
			utils.LogAll("could not get session for "+account.AccountId+":", err)
			continue
		}
		RemoveDocumentPermissions(sess, pointerAccountIds, documentName)
	}

	return nil
}

//Requires a list of account ids to remove from the document, and the documentName to remove the permissions from
func RemoveDocumentPermissions(sess *session.Session, accountIds []*string, documentName string) error {
	//The api has a limit of 20 account ids to remove at a time
	if len(accountIds) <= 20 {
		params := &ssm.ModifyDocumentPermissionInput{
			Name:               aws.String(documentName),
			AccountIdsToRemove: accountIds,
		}
		_, err := ssm.New(sess).ModifyDocumentPermission(params)
		if err != nil {
			return err
		}
	} else {
		var tempAccountIds []*string
		for i, accountId := range accountIds {
			tempAccountIds = append(tempAccountIds, accountId)
			if len(tempAccountIds) == 20 || (i+1) == len(accountIds) {
				params := &ssm.ModifyDocumentPermissionInput{
					Name:               aws.String(documentName),
					AccountIdsToRemove: tempAccountIds,
				}
				_, err := ssm.New(sess).ModifyDocumentPermission(params)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
