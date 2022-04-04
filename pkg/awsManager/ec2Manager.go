package awsManager

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/go-logr/logr"
	clientpkg "github.com/openshift/aws-account-shredder/pkg/aws"
	"github.com/openshift/aws-account-shredder/pkg/localMetrics"
)

const (
	maxBatchSize int = 50
)

// ListEc2InstancesForDeletion this lists all the instances that are eligible for deletion based on the tags and stored them in instances to be deleted
// this only creates an array of pointers and does not delete the instances
func ListEc2InstancesForDeletion(client clientpkg.Client, logger logr.Logger) []*string {

	var EC2InstancesToBeDeleted []*string
	token := ""
	for {
		ec2Descriptions, err := client.DescribeInstances(&ec2.DescribeInstancesInput{NextToken: aws.String(token)})
		if err != nil {
			logger.Error(err, "Failed to retrieve EC2 descriptions")
		}

		// nested for loop to read the tags , as it is a part of structure inside a structure output. Refer : https://pkg.go.dev/github.com/aws/aws-sdk-go/service/ec2?tab=doc#DescribeInstancesOutput
		for _, reservation := range ec2Descriptions.Reservations {
			for _, instance := range reservation.Instances {
				for _, tag := range instance.Tags {

					// If an EC2 instance matches the following conditions, store it for deletion
					if strings.HasPrefix(*tag.Key, "kubernetes.io") && (*instance.State.Code != 48) {
						EC2InstancesToBeDeleted = append(EC2InstancesToBeDeleted, instance.InstanceId)
						break
					}

					if (*tag.Key == "clusterAccountName" || *tag.Key == "clusterClaimLink" || *tag.Key == "clusterNamespace" || *tag.Key == "clusterClaimLinkNamespace") && (*instance.State.Code != 48) {
						EC2InstancesToBeDeleted = append(EC2InstancesToBeDeleted, instance.InstanceId)
						break
					}
				}
			}
		}

		// for pagination
		if ec2Descriptions.NextToken != nil {
			token = *ec2Descriptions.NextToken
		} else {
			break
		}
	}

	return EC2InstancesToBeDeleted
}

// DeleteEc2Instance deletes all ec2 instances in the given list
func DeleteEc2Instance(client clientpkg.Client, EC2InstancesToBeDeleted []*string, logger logr.Logger) error {
	if EC2InstancesToBeDeleted == nil {
		return nil
	}

	// We're batching the deletes to avoid hitting the limit of instances per request.
	// This uses a sliding window to iterate through the list in batches.
	totalToDelete := len(EC2InstancesToBeDeleted)
	for lowerBound, upperBound := 0, 0; lowerBound <= totalToDelete-1; lowerBound = upperBound {
		upperBound = lowerBound + maxBatchSize
		if upperBound > totalToDelete {
			upperBound = totalToDelete
		}
		batchedEC2Instances := EC2InstancesToBeDeleted[lowerBound:upperBound]
		_, err := client.TerminateInstances(&ec2.TerminateInstancesInput{InstanceIds: batchedEC2Instances})
		if err != nil {
			localMetrics.ResourceFail(localMetrics.Ec2Instance, client.GetRegion())
			logger.Error(err, "Failed to delete instances provided", "Instances", &batchedEC2Instances)
			return err
		}
	}
	localMetrics.ResourceSuccess(localMetrics.Ec2Instance, client.GetRegion())
	return nil
}

// CleanEc2Instances lists and deletes eligible ec2 instances
func CleanEc2Instances(client clientpkg.Client, logger logr.Logger) error {
	eC2InstancesToBeDeleted := ListEc2InstancesForDeletion(client, logger)
	err := DeleteEc2Instance(client, eC2InstancesToBeDeleted, logger)
	if err != nil {
		logger.Error(err, "Failed to delete ec2 instances")
		return err
	}
	logger.Info("All EC2 instances have been terminated for this region")
	return nil

}
