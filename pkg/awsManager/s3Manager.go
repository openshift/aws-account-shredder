package awsManager

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	clientpkg "github.com/openshift/aws-account-shredder/pkg/aws"
)

// this does not delete the S3 instances , this only creates an []* string for the resources that have to deleted
func ListS3InstancesForDeletion(client clientpkg.Client) []*string {

	var s3BucketsToBeDeleted []*string
	s3bucketDescription, err := client.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	for _, bucket := range s3bucketDescription.Buckets {
		s3BucketsToBeDeleted = append(s3BucketsToBeDeleted, bucket.Name)
	}

	return s3BucketsToBeDeleted
}

// this deletes the S3 buckets
func DeleteS3Buckets(client clientpkg.Client, s3BucketsToBeDeleted []*string) {

	for _, bucket := range s3BucketsToBeDeleted {

		// need to empty the bucket before the bucket can be deleted
		batchDeleteError := client.BatchDeleteBucketObjects(bucket)
		if batchDeleteError != nil {
			fmt.Println(batchDeleteError)
		}

		// Deleting the bucket
		_, err := client.DeleteBucket(&s3.DeleteBucketInput{Bucket: bucket})
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
}

func CleanS3Instances(client clientpkg.Client) {
	s3InstancesToBeDeleted := ListS3InstancesForDeletion(client)
	DeleteS3Buckets(client, s3InstancesToBeDeleted)
}
