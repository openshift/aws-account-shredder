package awsManager

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	clientpkg "github.com/openshift/aws-account-shredder/pkg/aws"
	"strings"
)

// this lists all the instances that are eligible for deletion based on the tags and stored them in instances to be deleted
// this only creates an array of pointers and does not delete the instances
func ListEc2InstancesForDeletion(client clientpkg.Client) []*string {

	var EC2InstancesToBeDeleted []*string
	token := ""
	for {
		ec2Descriptions, err := client.DescribeInstances(&ec2.DescribeInstancesInput{NextToken: aws.String(token)})
		if err != nil {
			fmt.Println("ERROR:", err)

		}

		// nested for loop to read the tags , as it is a part of structure inside a structure output. Refer : https://pkg.go.dev/github.com/aws/aws-sdk-go/service/ec2?tab=doc#DescribeInstancesOutput
		for _, reservation := range ec2Descriptions.Reservations {
			for _, instance := range reservation.Instances {
				for _, tag := range instance.Tags {

					if strings.HasPrefix(*tag.Key, "kubernetes.io") && (*instance.State.Code != 48) {

						EC2InstancesToBeDeleted = append(EC2InstancesToBeDeleted, instance.InstanceId)
						break
					}

					if (*tag.Key == "clusterAccountName" || *tag.Key == "clusterClaimLink" || *tag.Key == "clusterNamespace" || *tag.Key == "clusterClaimLinkNamespace") && (*instance.State.Code != 48) {
						EC2InstancesToBeDeleted = append(EC2InstancesToBeDeleted, instance.InstanceId)

						break
					}
					// no need to delete the instances if the above two 'if' conditions are not true

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

// this takes all the instances from insancesToBeDeleted  ( one by one ) and deletes them
func DeleteEc2Instance(client clientpkg.Client, EC2InstancesToBeDeleted []*string) {

	if EC2InstancesToBeDeleted == nil {
		fmt.Println("inside the if condition . it Hits ")
		return
	}

	_, err := client.TerminateInstances(&ec2.TerminateInstancesInput{InstanceIds: EC2InstancesToBeDeleted})
	if err != nil {
		if err, ok := err.(awserr.Error); ok {
			switch err.Code() {
			default:
				fmt.Println(err.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
	}
}

func CleanEc2Instances(client clientpkg.Client) {

	eC2InstancesToBeDeleted := ListEc2InstancesForDeletion(client)
	DeleteEc2Instance(client, eC2InstancesToBeDeleted)

}
