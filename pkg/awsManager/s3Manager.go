package awsManager

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-logr/logr"
	clientpkg "github.com/openshift/aws-account-shredder/pkg/aws"
	"github.com/openshift/aws-account-shredder/pkg/localMetrics"
)

//ListS3InstancesForDeletion creates a string list of s3 resources that need to be deleted
func ListS3InstancesForDeletion(client clientpkg.Client, logger logr.Logger) []*string {

	var s3BucketsToBeDeleted []*string
	s3bucketDescription, err := client.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		logger.Error(err, "Failed to list s3 buckets")
	}
	for _, bucket := range s3bucketDescription.Buckets {
		s3BucketsToBeDeleted = append(s3BucketsToBeDeleted, bucket.Name)
	}

	return s3BucketsToBeDeleted
}

//DeleteS3Buckets deletes the S3 buckets
// successful execution returns nil. Unsuccessful execution or errors occurred, would return an error
func DeleteS3Buckets(client clientpkg.Client, s3BucketsToBeDeleted []*string, logger logr.Logger) error {

	if s3BucketsToBeDeleted == nil {
		return nil
	}
	var s3BucketsNotDeleted []*string
	for _, bucket := range s3BucketsToBeDeleted {

		// need to empty the bucket before the bucket can be deleted
		batchDeleteError := client.BatchDeleteBucketObjects(bucket)
		if batchDeleteError != nil {
			logger.Error(batchDeleteError, "Failed to empty bucket", *bucket)
		}

		// Deleting the bucket
		_, err := client.DeleteBucket(&s3.DeleteBucketInput{Bucket: bucket})
		if err != nil {
			logger.Error(err, "could not delete bucket", *bucket)
			s3BucketsNotDeleted = append(s3BucketsNotDeleted, bucket)
			localMetrics.ResourceFail(localMetrics.S3Bucket, client.GetRegion())
			continue
		}
		localMetrics.ResourceSuccess(localMetrics.S3Bucket, client.GetRegion())
	}

	if s3BucketsNotDeleted != nil {
		return errors.New("s3BucketsNotDeleted")
	}

	return nil
}

// CleanS3Instances cleans s3 buckets
func CleanS3Instances(client clientpkg.Client, logger logr.Logger) error {
	s3InstancesToBeDeleted := ListS3InstancesForDeletion(client, logger)
	err := DeleteS3Buckets(client, s3InstancesToBeDeleted, logger)
	if err != nil {
		logger.Error(err, "Failed to delete s3 buckets")
		return err
	}
	logger.Info("All S3 buckets have been deleted for this region")
	return nil
}
