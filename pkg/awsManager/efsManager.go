package awsManager

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/go-logr/logr"
	clientpkg "github.com/openshift/aws-account-shredder/pkg/aws"
)

// CleanEFSMountTargets lists and then deletes listed efs mount targets
func CleanEFSMountTargets(client clientpkg.Client, logger logr.Logger) error {

	mountTargetToBeDeleted, err := listEFSMountTarget(client)
	if err != nil {
		logger.Error(err, "Failed to get list of EFS mount targets")
		return err
	}
	err = deleteEFSMountTarget(client, mountTargetToBeDeleted, logger)
	if err != nil {
		logger.Error(err, "Failed to delete mount targets")
		return err
	}

	logger.Info("Mount targets removed for this region")
	return nil
}

func listEFSMountTarget(client clientpkg.Client) ([]*string, error) {

	var marker *string
	var mountTargetsToBeDeleted []*string

	for {
		efsMounts, err := client.DescribeMountTargets(&efs.DescribeMountTargetsInput{Marker: marker})
		if err != nil {
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

func deleteEFSMountTarget(client clientpkg.Client, mountTargetToBeDeleted []*string, logger logr.Logger) error {

	var mountTargetNotDeleted []*string

	if mountTargetToBeDeleted == nil {
		return nil
	}

	for _, mountTarget := range mountTargetToBeDeleted {
		_, err := client.DeleteMountTarget(&efs.DeleteMountTargetInput{MountTargetId: mountTarget})
		if err != nil {
			logger.Error(err, "Unable to remove the mount-target", *mountTarget)
			mountTargetNotDeleted = append(mountTargetNotDeleted, mountTarget)
		}
	}

	if mountTargetNotDeleted != nil {
		return errors.New("FailedToRemoveAllMountTargets")
	}

	return nil
}

// CleanEFS lists and removes EFSs
func CleanEFS(client clientpkg.Client, logger logr.Logger) error {

	fileSystemToBeDeleted, err := listEFS(client, logger)
	if err != nil {
		logger.Error(err, "Failed to list EFS")
		return err
	}
	err = deleteEFS(client, fileSystemToBeDeleted, logger)
	if err != nil {
		logger.Error(err, "Failed to delete file systems")
		return err
	}

	logger.Info("all EFS removed for this region")
	return nil
}

func listEFS(client clientpkg.Client, logger logr.Logger) ([]*string, error) {

	var marker *string
	var filesystemToBeDeleted []*string

	for {
		fileSystemOutput, err := client.DescribeFileSystems(&efs.DescribeFileSystemsInput{Marker: marker})
		if err != nil {
			logger.Info("Can not list file system for this region")
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

func deleteEFS(client clientpkg.Client, fileSystemToBeDeleted []*string, logger logr.Logger) error {

	var fileSystemNotDeleted []*string

	if fileSystemToBeDeleted == nil {
		return nil
	}

	for _, fileSystem := range fileSystemToBeDeleted {
		_, err := client.DeleteFileSystem(&efs.DeleteFileSystemInput{FileSystemId: fileSystem})
		if err != nil {
			logger.Info("Unable to remove file system", *fileSystem)
			fileSystemNotDeleted = append(fileSystemNotDeleted, fileSystem)
		}
	}

	if fileSystemNotDeleted != nil {
		return errors.New("NotAllFileSystemsRemoved")
	}

	return nil
}
