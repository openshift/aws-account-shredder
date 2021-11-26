# aws-account-shredder
[![Go Report Card](https://goreportcard.com/badge/github.com/openshift/aws-account-shredder)](https://goreportcard.com/report/github.com/openshift/aws-account-shredder)
[![GoDoc](https://godoc.org/github.com/openshift/aws-account-shredder?status.svg)](https://pkg.go.dev/mod/github.com/openshift/aws-account-shredder)
[![License](https://img.shields.io/:license-apache-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)


Repository to audit, service, and clean up leftover AWS resources

## Prerequisites
* [osdctl](https://github.com/openshift/osdctl/) available in your `$PATH`
* a local kubernetes cluster ([crc](https://github.com/code-ready/crc/) or [kind](https://kind.sigs.k8s.io/))
* the Openshift Client [oc](https://github.com/openshift/oc)

## Deploying Shredder Locally

First make sure you have the following environment variables defined:
```
AWS_ACCESS_KEY_ID     # Access Key for the aws account that should be shredded
AWS_SECRET_ACCESS_KEY # Secret Key for the aws account that should be shredded
SHRED_ACCOUNT_ID      # ID of the Account that should be shredded
```

**Tip:** You can use [direnv](https://direnv.net) and add the above block (with variables filled in) into a `.envrc` file (make sure `.envrc` is in your global git ignore as well). Upon entry to the `aws-account-shredder` folder, the env vars inside the file will be loaded automatically, and unset when you leave the folder.

Load up CRC, Minishift or kind. Make sure you are logged in as an administrator. Then execute the pre-deploy step:

```
make predeploy
```

This will perform the following actions:
* create the `aws-account-operator` and `aws-account-shredder` namespaces
* Download and apply the `Account`-CRD from the [aws-account-operator](https://github.com/openshift/aws-account-operator/)
* use the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` to create the `aws-account-shreder-credentials`-secret.

> **Note:** If you find yourself changing your aws credentials while the shredder is running, you'll need to kill the shredder pod to ensure the updated secret is pulled when the new pod is created.

Assert that you have no failed accounts in the `aws-account-operator` namespace, otherwise these will be shredded once you run the next step:
```
oc -n aws-account-operator get accounts
```

Then, deploy the account shredder by running:
```
make deploy
```

## Running an ad-hoc shred

The following steps will perform the shredding on the account-id specified earlier. ( the `ACCOUNT_ID_TO_SHRED` environment variable ).

> DOUBLE CHECK THAT YOU'VE CONFIGURED THE CORRECT ACCOUNT ID, AS THIS IS A DESTRUCTIVE OPERATION THAT CAN NOT BE UNDONE!

Run the following to create the account:
```
make create-account
```

This will use the template [aws-shredder-account-delete-template.yaml](./hack/templates/aws-shredder-account-delete-template.yaml) to create a new account cr that represents your aws account. Then, use the next command to mark the account as `failed`, which will cause the account shredder to shred that account:

```
make shred-account
```

Once you set the status of the account to failed, the Shredder should pick it up and start shredding through the accounts.

You should be able to follow the logs and watch the shred happen using `make get-logs`.  Certain aws resources may not completely delete on the first attempt through the shredder, but the shredder will continue to run on the account until it is cleaned.

Once you are done with the cleanup, remove the Failed account (otherwise the shredder will infinitely loop over this account). You can accomplish this with `make delete-account`

To remove all created resources from your local cluster, you can run `make clean-operator`.

## Testing your changes locally

To test your changes locally, run `make test` and `make lint`. 

You'll need to have public repository to host the image you are going to build. For the following example we will maintain quay as the registry used.

You'll need to edit the `IMAGE_REPOSITORY` field in `project.mk` to point at your `aws-account-shredder` repo , not that of `app-sre`:
```
IMAGE_REPOSITORY?=<your username>
```
This is what's used by `standard.mk` when we run `make docker-build`. Alternatively, set the `IMAGE_REPOSITORY` as an environment variable:
```
export IMAGE_REPOSITORY=<your username>
```
So now that you're pointing to the correct repo, run:
```
make docker-build
```
This will build the container image and spit out a tag at the end, now we'll need to push this image up to the repo:
```
docker push quay.io/<username>/aws-account-shredder:<tag>
```

Now that you have your image pushed up, you can apply the deployment again using `make deploy`
