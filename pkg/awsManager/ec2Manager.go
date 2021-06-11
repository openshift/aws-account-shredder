package awsManager

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/go-logr/logr"
	clientpkg "github.com/openshift/aws-account-shredder/pkg/aws"
	"github.com/openshift/aws-account-shredder/pkg/localMetrics"
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

	_, err := client.TerminateInstances(&ec2.TerminateInstancesInput{InstanceIds: EC2InstancesToBeDeleted})
	if err != nil {
		if err, ok := err.(awserr.Error); ok {
			switch err.Code() {
			default:
				logger.Error(err, "Failed to delete instances provided", "Instances", &EC2InstancesToBeDeleted)
			}
		} else {
			logger.Error(err, "Failed to delete instances provided", "Instances", &EC2InstancesToBeDeleted)
		}
		localMetrics.ResourceFail(localMetrics.Ec2Instance, client.GetRegion())
		return errors.New("FailedToDeleteEc2Instance")
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
