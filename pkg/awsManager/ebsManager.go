package awsManager

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/go-logr/logr"
	clientpkg "github.com/openshift/aws-account-shredder/pkg/aws"
)

//ListEbsSnapshotForDeletion does not delete the Ebs snapshots, this only creates an []* string for the resources that have to deleted
func ListEbsSnapshotForDeletion(client clientpkg.Client, logger logr.Logger) []*string {

	var ebsSnapshotsToBeDeleted []*string
	var token *string
	// Filter only for snapshots owned by the account
	selfOwnerFilter := ec2.Filter{
		Name: aws.String("owner-alias"),
		Values: []*string{
			aws.String("self"),
		},
	}
	for {
		ebsSnapshotList, err := client.DescribeSnapshots(&ec2.DescribeSnapshotsInput{Filters: []*ec2.Filter{&selfOwnerFilter}, NextToken: token})
		if err != nil {
			logger.Error(err, "Failed to list EBS snapshots")
		}

		for _, ebsSnapshot := range ebsSnapshotList.Snapshots {
			ebsSnapshotsToBeDeleted = append(ebsSnapshotsToBeDeleted, ebsSnapshot.SnapshotId)
		}

		if ebsSnapshotList.NextToken != nil {
			token = ebsSnapshotList.NextToken
		} else {
			break
		}
	}

	return ebsSnapshotsToBeDeleted
}

// DeleteEbsSnapshots deletes the Ebs Snapshot
// successful execution returns nil. Unsuccessful execution or errors occured, would return an error
func DeleteEbsSnapshots(client clientpkg.Client, ebsSnapshotsToBeDeleted []*string, logger logr.Logger) error {

	if ebsSnapshotsToBeDeleted == nil {
		return nil
	}
	var ebsSnapshotsNotDeleted []*string
	for _, ebsSnapshotID := range ebsSnapshotsToBeDeleted {

		_, ebsSnapshotDeleteError := client.DeleteSnapshot(&ec2.DeleteSnapshotInput{SnapshotId: ebsSnapshotID})
		if ebsSnapshotDeleteError != nil {
			logger.Error(ebsSnapshotDeleteError, "Failed to delete snapshot", *ebsSnapshotID)
			ebsSnapshotsNotDeleted = append(ebsSnapshotsNotDeleted, ebsSnapshotID)
		}
	}

	if ebsSnapshotsNotDeleted != nil {
		return errors.New("FailedComprehensiveSnapshotDeletion")
	}

	return nil
}

func listVolumeForDeletion(client clientpkg.Client, logger logr.Logger) []*string {

	var token *string
	var ebsVolumesToBeDeleted []*string

	for {
		ebsVolumeList, err := client.DescribeVolumes(&ec2.DescribeVolumesInput{NextToken: token})
		if err != nil {
			logger.Error(err, "Failed to retrieve Volume list")
			return nil
		}

		for _, ebsVolume := range ebsVolumeList.Volumes {

			if *ebsVolume.State == "available" {
				ebsVolumesToBeDeleted = append(ebsVolumesToBeDeleted, ebsVolume.VolumeId)
			}
		}

		if ebsVolumeList.NextToken != nil {
			token = ebsVolumeList.NextToken
		} else {
			break
		}
	}
	return ebsVolumesToBeDeleted
}

func deleteEbsVolumes(client clientpkg.Client, ebsVolumesToBeDeleted []*string, logger logr.Logger) error {

	if ebsVolumesToBeDeleted == nil {
		return nil
	}
	var ebsVolumesNotDeleted []*string
	for _, ebsVolumeID := range ebsVolumesToBeDeleted {

		_, err := client.DeleteVolume(&ec2.DeleteVolumeInput{VolumeId: ebsVolumeID})
		if err != nil {
			logger.Error(err, "Failed to delete Volume", *ebsVolumeID)
			ebsVolumesNotDeleted = append(ebsVolumesNotDeleted, ebsVolumeID)
		}
	}

	if ebsVolumesNotDeleted != nil {
		return errors.New("FailedComprehensiveEBSVolumesDeletion")
	}

	return nil
}

// CleanEbsSnapshots lists and deletes EBS Snapshots
func CleanEbsSnapshots(client clientpkg.Client, logger logr.Logger) error {
	ebsSnapshotsToBeDeleted := ListEbsSnapshotForDeletion(client, logger)
	err := DeleteEbsSnapshots(client, ebsSnapshotsToBeDeleted, logger)
	if err != nil {
		logger.Error(err, "Failed to delete EBS snapshots")
		return err
	}
	logger.Info("All EBS snapshots have been removed for this region")
	return nil
}

// CleanEbsVolumes lists and deletes EBS volumes
func CleanEbsVolumes(client clientpkg.Client, logger logr.Logger) error {
	ebsVolumeToBeDeleted := listVolumeForDeletion(client, logger)
	err := deleteEbsVolumes(client, ebsVolumeToBeDeleted, logger)
	if err != nil {
		logger.Error(err, "Failed to delete EBS volumes")
		return err
	}
	logger.Info("All EBS volumes have been removed successfully for this region")
	return nil
}
