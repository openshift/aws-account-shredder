package main

import (
	"fmt"
	_ "fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
	clientpkg "github.com/openshift/aws-account-shredder/pkg/aws"
	"os"
)

const (

	// the credentials for aws-account-shredder
	accessID    = ""
	secretKey   = ""
	accountID   = "" // it is 12 digit account id , the sub level account , for future use and debugging purpose
	sessionName = ""
)

var (
	supportedRegions = []string{"us-east-1", "us-east-2", "us-west-1", "us-west-2", "ca-central-1", "eu-central-1", "eu-west-1", "eu-west-2", "eu-west-3", "ap-northeast-1", "ap-northeast-2", "ap-south-1", "ap-southeast-1", "ap-southeast-2", "sa-east-1"}
)

func main() {

	// creating a new cient with us-east-1 region by default
	client, err := clientpkg.NewClient(accessID, secretKey, "", "us-east-1")
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	RoleArnParameter := "arn:aws:iam::" + accountID + ":role/OrganizationAccountAccessRole"
	assumedRole, err := client.AssumeRole(&sts.AssumeRoleInput{RoleArn: aws.String(RoleArnParameter), RoleSessionName: aws.String(sessionName)})
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
	assumedAccessKey := *assumedRole.Credentials.AccessKeyId
	assumedSecretKey := *assumedRole.Credentials.SecretAccessKey
	assumedSessionToken := *assumedRole.Credentials.SessionToken

	// for debugging purpose only.
	//	fmt.Println("new access id : ", assumedAccessKey)
	//	fmt.Println("new secret key is", aithub
	//	ssumedSecretKey)
	//	fmt.Println("new session token\n\n", assumedSessionToken)

	// looping through all the regions

	for _, region := range supportedRegions {
		fmt.Println("\n EC2 instances in region ", region)
		assumedRoleClient, err := clientpkg.NewClient(assumedAccessKey, assumedSecretKey, assumedSessionToken, region)
		if err != nil {
			fmt.Println("ERROR:", err)
			os.Exit(1)
		}
		// just for debugging purpose
		//fmt.Println("new_client is ", client2)

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

}
