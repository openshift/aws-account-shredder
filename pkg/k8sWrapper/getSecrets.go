package k8sWrapper

import (
	"context"
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	awsCredsSecretIDKey = "aws_access_key_id"
	// #nosec - G101 no hardcoded credentials
	awsCredsSecretAccessKey = "aws_secret_access_key"
	namespace               = "aws-account-shredder" // change the namespace according to your environment. this is the namespace, from where secret has to retreived from
	// #nosec - G101 no hardcoded credentials
	secretName = "aws-account-shredder-credentials" // the name of the secret to be read
)

// read the credentials stored in aws-account-shredder to start up the connection to AWS
func GetAWSAccountCredentials(ctx context.Context, cli client.Client) (string, string, error) {

	var secret v1.Secret

	if err := cli.Get(ctx, types.NamespacedName{
		Name:      secretName,
		Namespace: namespace,
	}, &secret); err != nil {
		return "", "", err
	}
	accessKeyID, ok := secret.Data[awsCredsSecretIDKey]
	if !ok {
		fmt.Println("AWS Access Key ID could not be decoded")
		return "", "", errors.New("ERROR : AWS Access Key ID could not be decoded ")
	}
	secretAccessKey, ok := secret.Data[awsCredsSecretAccessKey]
	if !ok {
		fmt.Println("AWS access secret key could not be decoded")
		return "", "", errors.New("ERROR: AWS access secret key could not be decoded ")
	}

	return string(accessKeyID), string(secretAccessKey), nil
}
