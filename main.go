package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/openshift/aws-account-operator/pkg/apis/aws/v1alpha1"
	clientpkg "github.com/openshift/aws-account-shredder/pkg/aws"
	"github.com/openshift/aws-account-shredder/pkg/awsManager"
	"github.com/openshift/aws-account-shredder/pkg/awsv1alpha1"
	"github.com/openshift/aws-account-shredder/pkg/k8sWrapper"
	"github.com/openshift/aws-account-shredder/pkg/localMetrics"
	"github.com/openshift/operator-custom-metrics/pkg/metrics"
	clientGoScheme "k8s.io/client-go/kubernetes/scheme"
	kubeRest "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	sessionName = "awsAccountShredder"
	metricsPort = "8080"
	metricsPath = "/metrics"
)

var (
	supportedRegions = []string{"us-east-1", "us-east-2", "us-west-1", "us-west-2", "ca-central-1", "eu-central-1", "eu-west-1", "eu-west-2", "eu-west-3", "ap-northeast-1", "ap-northeast-2", "ap-south-1", "ap-southeast-1", "ap-southeast-2", "sa-east-1"}
)

func main() {
	// creates the in-cluster config
	config, err := kubeRest.InClusterConfig()
	if err != nil {
		fmt.Println(err)
	}

	// creating a client for reading the AccountID
	cli, err := client.New(config, client.Options{})
	if err != nil {
		fmt.Println("ERROR:", err)
	}

	//integrating the account CRD to this project
	v1alpha1.AddToScheme(clientGoScheme.Scheme)

	if err := routev1.AddToScheme(clientGoScheme.Scheme); err != nil {
		fmt.Println("ERROR: ", err)
	}

	//Create localMetrics endpoint and register localMetrics
	metricsServer := metrics.NewBuilder().WithPort(metricsPort).WithPath(metricsPath).
		WithCollectors(localMetrics.MetricsList).
		WithRoute().
		WithServiceName("aws-account-shredder").
		GetConfig()

	// Configure localMetrics if it errors log the error but continue
	if err := metrics.ConfigureMetrics(context.TODO(), *metricsServer); err != nil {
		fmt.Println(err, "Failed to configure Metrics")

	}

	//reading the aws-account-shredder-credentials secret
	accessKeyID, secretAccessKey, err := k8sWrapper.GetAWSAccountCredentials(context.TODO(), cli)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}

	// creating a new AWSclient with the information extracted from the secret file
	awsClient, err := clientpkg.NewClient(accessKeyID, secretAccessKey, "", "us-east-1")
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	for {
		// reading the account ID to be cleared
		accountIDlist, err := awsv1alpha1.GetAccountIDsToReset(context.TODO(), cli)
		if err != nil {
			fmt.Println("ERROR: ", err)
		}

		for _, accountID := range accountIDlist {
			fmt.Println("Now Processing AccountID: ", accountID)

			// assuming roles for the given AccountID
			RoleArnParameter := "arn:aws:iam::" + accountID + ":role/OrganizationAccountAccessRole"
			assumedRole, err := awsClient.AssumeRole(&sts.AssumeRoleInput{RoleArn: aws.String(RoleArnParameter), RoleSessionName: aws.String(sessionName)})
			if err != nil {
				fmt.Println("ERROR:", err)
				// need continue , or else the next line will throw an error ( non existing pointer being deferenced)
				// hence moving on to next element
				continue

			}
			assumedAccessKey := *assumedRole.Credentials.AccessKeyId
			assumedSecretKey := *assumedRole.Credentials.SecretAccessKey
			assumedSessionToken := *assumedRole.Credentials.SessionToken

			for _, region := range supportedRegions {
				fmt.Println("\n Current Region : ", region)
				assumedRoleClient, err := clientpkg.NewClient(assumedAccessKey, assumedSecretKey, assumedSessionToken, region)
				if err != nil {
					fmt.Println("ERROR:", err)
				}

				awsManager.CleanS3Instances(assumedRoleClient)
				awsManager.CleanEc2Instances(assumedRoleClient)
				awsManager.ClearRoute53Resources(assumedRoleClient)
				awsManager.CleanUpAwsRoute53(assumedRoleClient)
			}
		}
	}
}
