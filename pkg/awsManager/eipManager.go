package awsManager

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/go-logr/logr"
	clientpkg "github.com/openshift/aws-account-shredder/pkg/aws"
)

// CleanEIPAddresses Cleans any hanging EIPAddresses
func CleanEIPAddresses(client clientpkg.Client, logger logr.Logger) error {
	result, err := client.DescribeAddresses(&ec2.DescribeAddressesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("domain"),
				Values: aws.StringSlice([]string{"vpc"}),
			},
		},
	})
	if err != nil {
		logger.Error(err, "Unable to get elastic IP address")
		return err
	}

	// Release the IP addresses if there are any.
	if len(result.Addresses) == 0 {
		logger.Info("No elastic IPs for current region")
	} else {
		// Loop through all EIP addresses
		for _, address := range result.Addresses {
			logger.Info("Attempting to release EIP address", "allocationID", address.AllocationId)
			err := realeaseEIPAddress(client, logger, *address.AllocationId)
			if err != nil {
				return err
			}
		}
		logger.Info("Successfully released all EIP addresses in the current region")
	}
	return nil
}

func realeaseEIPAddress(client clientpkg.Client, logger logr.Logger, allocationID string) error {
	_, err := client.ReleaseAddress(&ec2.ReleaseAddressInput{
		AllocationId: aws.String(allocationID),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "InvalidAllocationID.NotFound" {
			logger.Error(err, "Allocation ID does not exist", "allocationID", allocationID)
		}
		logger.Error(err, "Unable to release IP address for allocation", "allocationID", allocationID)
		return err
	}
	logger.Info("Successfully released allocation ID", "allocationID", allocationID)
	return nil
}
