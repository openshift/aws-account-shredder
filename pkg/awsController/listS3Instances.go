package awsController

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	clientpkg "github.com/openshift/aws-account-shredder/pkg/aws"
	"os"
)

func ListS3Instances(assumedRoleClient clientpkg.Client) {

	s3bucketDescription, err := assumedRoleClient.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
	fmt.Println("S3 instances", s3bucketDescription)

}
