package ssm

import (
	"fmt"
	"github.com/afeeblechild/aws-go-tool/lib/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"log"
)

type Baseline struct {
	OS           string
	ApproveDelay int64
}

//AddRegionBaseline will add a single baseline, and infer the rules based on the baseline input
func AddRegionBaseline(sess *session.Session, baseline Baseline) (string, error) {
	ruleGroup := &ssm.PatchRuleGroup{
		PatchRules: []*ssm.PatchRule{
			{
				PatchFilterGroup: &ssm.PatchFilterGroup{
					PatchFilters: []*ssm.PatchFilter{
						{
							Key:    aws.String("SEVERITY"),
							Values: aws.StringSlice([]string{"Critical", "Important"}),
						},
						{
							Key:    aws.String("CLASSIFICATION"),
							Values: aws.StringSlice([]string{"Security"}),
						},
					},
				},
				ApproveAfterDays: aws.Int64(baseline.ApproveDelay),
			},
		},
	}

	name := "2W-" + baseline.OS + "-PatchBaseline"
	description := "Baseline for 2W " + baseline.OS + " patching"
	params := &ssm.CreatePatchBaselineInput{
		Name:            aws.String(name),
		Description:     aws.String(description),
		ApprovalRules:   ruleGroup,
		OperatingSystem: aws.String(baseline.OS),
	}
	if baseline.OS != "WINDOWS" {
		params.RejectedPatches = aws.StringSlice([]string{"*kernel*"})
	}
	if baseline.OS == "WINDOWS" {
		ruleGroup := &ssm.PatchRuleGroup{
			PatchRules: []*ssm.PatchRule{
				{
					PatchFilterGroup: &ssm.PatchFilterGroup{
						PatchFilters: []*ssm.PatchFilter{
							{
								Key:    aws.String("MSRC_SEVERITY"),
								Values: aws.StringSlice([]string{"Critical", "Important"}),
							},
							{
								Key:    aws.String("CLASSIFICATION"),
								Values: aws.StringSlice([]string{"SecurityUpdates", "CriticalUpdates"}),
							},
						},
					},
					ApproveAfterDays: aws.Int64(baseline.ApproveDelay),
				},
			},
		}
		params.ApprovalRules = ruleGroup
	}
	if baseline.OS == "UBUNTU" {
		ruleGroup := &ssm.PatchRuleGroup{
			PatchRules: []*ssm.PatchRule{
				{
					PatchFilterGroup: &ssm.PatchFilterGroup{
						PatchFilters: []*ssm.PatchFilter{
							{
								Key:    aws.String("PRIORITY"),
								Values: aws.StringSlice([]string{"Required"}),
							},
						},
					},
					ApproveAfterDays: aws.Int64(baseline.ApproveDelay),
				},
			},
		}
		params.ApprovalRules = ruleGroup
	}

	resp, err := ssm.New(sess).CreatePatchBaseline(params)
	if err != nil {
		return "", err
	}
	return *resp.BaselineId, nil
}

func AddDefaultBaseline(sess *session.Session, id string) error {
	params := &ssm.RegisterDefaultPatchBaselineInput{
		BaselineId: aws.String(id),
	}
	_, err := ssm.New(sess).RegisterDefaultPatchBaseline(params)
	if err != nil {
		return err
	}
	return nil
}

func AddRegionBaselines(OSList []string, account utils.AccountInfo) {
	fmt.Println("Creating Baselines for", account.Profile)
	for _, OS := range OSList {
		sess, err := account.GetSession()
		if err != nil {
			log.Println("could not open session in", account.Region, "for", account.Profile)
			continue
		}
		baseline := Baseline{
			OS:           OS,
			ApproveDelay: 3,
		}
		id, err := AddRegionBaseline(sess, baseline)
		if err != nil {
			log.Println("could not add baseline", OS, ":", err)
		}

		err = AddDefaultBaseline(sess, id)
		if err != nil {
			log.Println("could not add baseline", OS, "as default:", err)
		}
	}
}
