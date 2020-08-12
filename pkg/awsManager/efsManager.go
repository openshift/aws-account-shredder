package awsManager

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/service/efs"
	clientpkg "github.com/openshift/aws-account-shredder/pkg/aws"
)

func CleanEFSMountTargets(client clientpkg.Client) error {

	mountTargetToBeDeleted, err := listEFSMountTarget(client)
	if err != nil {
		return err
	}
	err = deleteEFSMountTarget(client, mountTargetToBeDeleted)
	if err != nil {
		return err
	}

	fmt.Print("Mount targets removed for this region")
	return nil
}

func listEFSMountTarget(client clientpkg.Client) ([]*string, error) {

	var marker *string
	var mountTargetsToBeDeleted []*string

	for {
		efsMounts, err := client.DescribeMountTargets(&efs.DescribeMountTargetsInput{Marker: marker})
		if err != nil {
			fmt.Println("Can not list the mount target for this region")
			return nil, err
		}

		for _, mountTarget := range efsMounts.MountTargets {
			mountTargetsToBeDeleted = append(mountTargetsToBeDeleted, mountTarget.MountTargetId)
		}

		if efsMounts.NextMarker != nil {
			marker = efsMounts.NextMarker
		} else {
			break
		}
	}

	return mountTargetsToBeDeleted, nil

}

func deleteEFSMountTarget(client clientpkg.Client, mountTargetToBeDeleted []*string) error {

	var mountTargetNotDeleted []*string

	if mountTargetToBeDeleted == nil {
		return nil
	}

	for _, mountTarget := range mountTargetToBeDeleted {
		_, err := client.DeleteMountTarget(&efs.DeleteMountTargetInput{MountTargetId: mountTarget})
		if err != nil {
			fmt.Print("Unable to remove the mount-target", *mountTarget)
			mountTargetNotDeleted = append(mountTargetNotDeleted, mountTarget)
		}
	}

	if mountTargetNotDeleted != nil {
		return errors.New("not all mount targets were removed")
	}

	return nil
}

func CleanEFS(client clientpkg.Client) error {

	fileSystemToBeDeleted, err := listEFS(client)
	if err != nil {
		return err
	}
	err = deleteEFS(client, fileSystemToBeDeleted)
	if err != nil {
		return err
	}

	fmt.Print("all EFS removed for this region")
	return nil
}

func listEFS(client clientpkg.Client) ([]*string, error) {

	var marker *string
	var filesystemToBeDeleted []*string

	for {
		fileSystemOutput, err := client.DescribeFileSystems(&efs.DescribeFileSystemsInput{Marker: marker})
		if err != nil {
			fmt.Println("Can not list file system for this region")
			return nil, err
		}

		for _, fileSystem := range fileSystemOutput.FileSystems {
			filesystemToBeDeleted = append(filesystemToBeDeleted, fileSystem.FileSystemId)
		}

		if fileSystemOutput.NextMarker != nil {
			marker = fileSystemOutput.NextMarker
		} else {
			break
		}

	}

	return filesystemToBeDeleted, nil
}

func deleteEFS(client clientpkg.Client, fileSystemToBeDeleted []*string) error {

	var fileSystemNotDeleted []*string

	if fileSystemToBeDeleted == nil {
		return nil
	}

	for _, fileSystem := range fileSystemToBeDeleted {
		_, err := client.DeleteFileSystem(&efs.DeleteFileSystemInput{FileSystemId: fileSystem})
		if err != nil {
			fmt.Print("Unable to remove file system", *fileSystem)
			fileSystemNotDeleted = append(fileSystemNotDeleted, fileSystem)
		}
	}

	if fileSystemNotDeleted != nil {
		return errors.New("not all file system were removed")
	}

	return nil
}
