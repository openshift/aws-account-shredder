package awsv1alpha1

import (
	"context"
	"fmt"

	awsv1alpha1 "github.com/openshift/aws-account-operator/pkg/apis/aws/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// read the account ID and pass it to the cleanup service
func GetAccountCRsToReset(ctx context.Context, cli client.Client) ([]awsv1alpha1.Account, error) {

	var accounts awsv1alpha1.AccountList
	err := cli.List(ctx, &accounts, &client.ListOptions{
		Namespace: "aws-account-operator",
	})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var accountCRs []awsv1alpha1.Account
	for _, account := range accounts.Items {
		if account.Spec.ClaimLink == "" && account.Status.State == "Failed" && !account.Spec.BYOC {
			accountCRs = append(accountCRs, account)
		}

	}
	return accountCRs, err
}
