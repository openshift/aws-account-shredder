package awsManager

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	clientpkg "github.com/openshift/aws-account-shredder/pkg/aws"
	"os"
)

func ListEc2Instances(assumedRoleClient clientpkg.Client) {

	token := ""
	for {
		ec2Descriptions, err := assumedRoleClient.DescribeInstances(&ec2.DescribeInstancesInput{NextToken: aws.String(token)})
		if err != nil {
			fmt.Println("ERROR:", err)
			os.Exit(1)
		}
		fmt.Println("EC2 instances", ec2Descriptions)
		if ec2Descriptions.NextToken != nil {
			token = *ec2Descriptions.NextToken
		} else {
			break
		}
	}

}
