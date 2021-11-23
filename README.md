# aws-account-shredder
[![Go Report Card](https://goreportcard.com/badge/github.com/openshift/aws-account-shredder)](https://goreportcard.com/report/github.com/openshift/aws-account-shredder)
[![GoDoc](https://godoc.org/github.com/openshift/aws-account-shredder?status.svg)](https://pkg.go.dev/mod/github.com/openshift/aws-account-shredder)
[![License](https://img.shields.io/:license-apache-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)


Repository to audit, service, and clean up leftover AWS resources

## Deploying Shredder Locally

First make sure you have the following environment variables defined:
```
AWS_ACCESS_KEY_ID     # Access Key for the aws account that should be shredded
AWS_SECRET_ACCESS_KEY # Secret Key for the aws account that should be shredded
SHRED_ACCOUNT_ID      # ID of the Account that should be shredded
```

Load up CRC, Minishift or kind. Make sure you are logged in as an administrator. Then execute the pre-deploy step:

```
$> make predeploy
```

This will perform the following actions:
* create the `aws-account-operator`-namespace
* Download and apply the `Account`-crd and service-account declarations from the [aws-account-operator](https://github.com/openshift/aws-account-operator/)
* use the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` to create the `aws-account-shreder-credentials`-secret.

> **Note:** If you find yourself changing your aws credentials while the shredder is running, you'll need to kill the shredder pod to ensure the updated secret is pulled when the new pod is created.

Assert that you have no failed accounts in the `aws-account-operator` namespace, otherwise these will be shredded once you run the next step:
```
$> oc -n aws-account-operator get accounts
```

Then, deploy the account shredder by running:
```
$> make deploy
```

## Running an ad-hoc shred

The following steps will perform the shredding on the account-id specified earlier. ( the `SHRED_ACCOUNT_ID` environment variable ).

> DOUBLE CHECK THAT YOU'VE CONFIGURED THE CORRECT ACCOUNT ID, AS THIS IS A DESTRUCTIVE OPERATION THAT CAN NOT BE UNDONE!

Run the following to create the account:
```
make create-account
```

This will use the template [aws-account-shredder-delete.yaml.tpl](./hack/templates/aws-shredder-account-delete.yaml.tpl) to create a new account cr that represents your aws account. Then, use the next command to mark the account as `failed`, which will cause the account shredder to shred that account:

```
make shred-account
```
You should be able to follow the logs and watch the shred happen using `make get-logs`.  Certain objects may not delete on the first attempt through the shredder, but the shredder will continue to run on the account until it is created.

Once you are done with the cleanup, remove the Failed account (otherwise the shredder will infinitely loop over this account). You can accomplish this with `make delete-account`.

## Testing your changes locally

To test your changes locally, you'll need to have public repository to host the image you are going to build. For the following example we will maintain quay as the registry used.

You'll need to edit the `IMAGE_REPOSITORY` field in `project.mk` to point at your `aws-account-shredder` repo , not that of `app-sre`:
```
IMAGE_REPOSITORY?=<your username>
```
This is what's used by `standard.mk` when we run `make docker-build`. So now that you're pointing to the correct repo, run:
```
make docker-build
```
This will build the container image and spit out a tag at the end, now we'll need to push this image up to the repo:
```
docker push quay.io/<username>/aws-account-shredder:<tag>
```
You'll now need to update `deploy/deployment.yaml` to point at the image you just built:
```
image: quay.io/<username>/aws-account-shredder:<image tag>
```
Now that you have the your image pushed up and your deployment updated, you can simply apply the deployment to roll out your local test branch.
