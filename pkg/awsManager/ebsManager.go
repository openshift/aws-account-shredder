package awsManager

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	clientpkg "github.com/openshift/aws-account-shredder/pkg/aws"
)

// this does not delete the Ebs snapshots, this only creates an []* string for the resources that have to deleted
func ListEbsSnapshotForDeletion(client clientpkg.Client) []*string {

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
			fmt.Println("ERROR:", err)
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

// this deletes the Ebs Snapshot
// successful execution returns nil. Unsuccessful execution or errors occured, would return an error
func DeleteEbsSnapshots(client clientpkg.Client, ebsSnapshotsToBeDeleted []*string) error {

	if ebsSnapshotsToBeDeleted == nil {
		return nil
	}
	var ebsSnapshotsNotDeleted []*string
	for _, ebsSnapshotId := range ebsSnapshotsToBeDeleted {

		_, ebsSnapshotDeleteError := client.DeleteSnapshot(&ec2.DeleteSnapshotInput{SnapshotId: ebsSnapshotId})
		if ebsSnapshotDeleteError != nil {
			fmt.Println(ebsSnapshotDeleteError)
			fmt.Print("Could not delete snapshot :", *ebsSnapshotId)
			ebsSnapshotsNotDeleted = append(ebsSnapshotsNotDeleted, ebsSnapshotId)
		}
	}

	if ebsSnapshotsNotDeleted != nil {
		return errors.New("Could not delete all snapshots for this region")
	}

	return nil
}

func ListVolumeForDeletion(client clientpkg.Client) []*string {

	var token *string
	var ebsVolumesToBeDeleted []*string

	for {
		ebsVolumeList, err := client.DescribeVolumes(&ec2.DescribeVolumesInput{NextToken: token})
		if err != nil {
			fmt.Print(err)
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

func DeleteEbsVolumes(client clientpkg.Client, ebsVolumesToBeDeleted []*string) error {

	if ebsVolumesToBeDeleted == nil {
		return nil
	}
	var ebsVolumesNotDeleted []*string
	for _, ebsVolumeId := range ebsVolumesToBeDeleted {

		_, err := client.DeleteVolume(&ec2.DeleteVolumeInput{VolumeId: ebsVolumeId})
		if err != nil {
			fmt.Println(err)
			fmt.Print("Could not delete Volume :", *ebsVolumeId)
			ebsVolumesNotDeleted = append(ebsVolumesNotDeleted, ebsVolumeId)
		}
	}

	if ebsVolumesNotDeleted != nil {
		return errors.New("Could not delete all EBS volume for this region")
	}

	return nil
}

func CleanEbsSnapshots(client clientpkg.Client) error {
	ebsSnapshotsToBeDeleted := ListEbsSnapshotForDeletion(client)
	err := DeleteEbsSnapshots(client, ebsSnapshotsToBeDeleted)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("All EBS snapshots have been removed for this region")
	return nil
}

func CleanEbsVolumes(client clientpkg.Client) error {

	ebsVolumeToBeDeleted := ListVolumeForDeletion(client)
	err := DeleteEbsVolumes(client, ebsVolumeToBeDeleted)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("All EBS volumes have been removed successfully for this region")
	return nil
}
