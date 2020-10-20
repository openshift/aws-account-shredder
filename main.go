package main

import (
	"context"

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
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/prometheus/client_golang/prometheus"
	clientGoScheme "k8s.io/client-go/kubernetes/scheme"
	kubeRest "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	sessionName = "awsAccountShredder"
	metricsPort = "8080"
	metricsPath = "/metrics"
)

var (
	supportedRegions = []string{"us-east-1", "us-east-2", "us-west-1", "us-west-2", "ca-central-1", "eu-central-1", "eu-west-1", "eu-west-2", "eu-west-3", "ap-northeast-1", "ap-northeast-2", "ap-south-1", "ap-southeast-1", "ap-southeast-2", "sa-east-1"}
	log              = logf.Log.WithName("shredder_logger")
)

func main() {
	logf.SetLogger(zap.Logger())

	// creates the in-cluster config
	config, err := kubeRest.InClusterConfig()
	if err != nil {
		log.Error(err, "Failed to retrieve in cluster config")
	}

	// creating a client for reading the AccountID
	cli, err := client.New(config, client.Options{})
	if err != nil {
		log.Error(err, "Failed to initialize new client")
	}

	//integrating the account CRD to this project
	v1alpha1.AddToScheme(clientGoScheme.Scheme)

	if err := routev1.AddToScheme(clientGoScheme.Scheme); err != nil {
		log.Error(err, "Failed to integrate account CR to scheme")
	}

	//Create localMetrics endpoint and register localMetrics
	var collectors = []prometheus.Collector{}
	for _, collector := range localMetrics.MetricsList {
		collectors = append(collectors, collector)
	}
	metricsServer := metrics.NewBuilder().WithPort(metricsPort).WithPath(metricsPath).
		WithCollectors(collectors).
		WithRoute().
		WithServiceName("aws-account-shredder").
		GetConfig()

	// Configure localMetrics if it errors log the error but continue
	if err := metrics.ConfigureMetrics(context.TODO(), *metricsServer); err != nil {
		log.Error(err, "Failed to configure metrics")
	}

	//reading the aws-account-shredder-credentials secret
	accessKeyID, secretAccessKey, err := k8sWrapper.GetAWSAccountCredentials(context.TODO(), cli)
	if err != nil {
		log.Error(err, "Failed to read aws-account-shredder-credentials secret")
	}

	// creating a new AWSclient with the information extracted from the secret file
	awsClient, err := clientpkg.NewClient(accessKeyID, secretAccessKey, "", "us-east-1")
	if err != nil {
		log.Error(err, "Failed to create new AWSclient")
	}
	for {
		// reading the account ID to be cleared
		accountCRList, err := awsv1alpha1.GetAccountCRsToReset(context.TODO(), cli)
		if err != nil {
			log.Error(err, "Failed to retieve list of accounts to clean")
		}

		for _, account := range accountCRList {
			logger := log.WithValues("AccountName", account.Name, "AccountID", account.Spec.AwsAccountID)
			logger.Info("New Account being shredder") // Usefull for keeping track of when work begins on an account
			// assuming roles for the given AccountID
			RoleArnParameter := "arn:aws:iam::" + account.Spec.AwsAccountID + ":role/OrganizationAccountAccessRole"
			assumedRole, err := awsClient.AssumeRole(&sts.AssumeRoleInput{RoleArn: aws.String(RoleArnParameter), RoleSessionName: aws.String(sessionName)})
			if err != nil {
				logger.Error(err, "Failed to assume necessary account role", RoleArnParameter)
				// need continue , or else the next line will throw an error ( non existing pointer being deferenced)
				// hence moving on to next element
				continue

			}
			assumedAccessKey := *assumedRole.Credentials.AccessKeyId
			assumedSecretKey := *assumedRole.Credentials.SecretAccessKey
			assumedSessionToken := *assumedRole.Credentials.SessionToken

			var allErrors []error
			for _, region := range supportedRegions {
				logger = log.WithValues("AccountName", account.Name, "AccountID", account.Spec.AwsAccountID, "Region", region)
				assumedRoleClient, err := clientpkg.NewClient(assumedAccessKey, assumedSecretKey, assumedSessionToken, region)
				if err != nil {
					logger.Error(err, "Failed to initialize new AWS client")
				}
				allErrors = append(allErrors, awsManager.CleanS3Instances(assumedRoleClient, logger))
				allErrors = append(allErrors, awsManager.CleanEc2Instances(assumedRoleClient, logger))
				allErrors = append(allErrors, awsManager.CleanUpAwsRoute53(assumedRoleClient, logger))
				allErrors = append(allErrors, awsManager.CleanEFSMountTargets(assumedRoleClient, logger))
				allErrors = append(allErrors, awsManager.CleanEFS(assumedRoleClient, logger))
				allErrors = append(allErrors, awsManager.CleanVpcInstances(assumedRoleClient, logger))
				allErrors = append(allErrors, awsManager.CleanEbsSnapshots(assumedRoleClient, logger))
				allErrors = append(allErrors, awsManager.CleanEbsVolumes(assumedRoleClient, logger))
			}
			// After cleaning up every region we set the account state to Ready if no errors were encountered
			resetAccount := true
			for _, err := range allErrors {
				if err != nil {
					resetAccount = false
				}
			}
			if resetAccount {
				awsv1alpha1.SetAccountStateReady(cli, account)
			}
		}
	}
}
