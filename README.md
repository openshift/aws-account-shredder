# aws-account-shredder
[![Go Report Card](https://goreportcard.com/badge/github.com/openshift/aws-account-shredder)](https://goreportcard.com/report/github.com/openshift/aws-account-shredder)
[![GoDoc](https://godoc.org/github.com/openshift/aws-account-shredder?status.svg)](https://pkg.go.dev/mod/github.com/openshift/aws-account-shredder)
[![License](https://img.shields.io/:license-apache-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)


Repository to audit, service, and clean up leftover AWS resources. 

Once deployed, the AWS Account Shredder runs continuously looking for Account CR's in the `aws-account-operator` namespace
with a "Failed" state. Any such Account CR's will have their associated AWS resources cleaned up before resetting the Account CR
state.

## Prerequisites
* [osdctl](https://github.com/openshift/osdctl/) available in your `$PATH`
* a local kubernetes cluster ([crc](https://github.com/code-ready/crc/) or [kind](https://kind.sigs.k8s.io/))
* the Openshift Client [oc](https://github.com/openshift/oc)

## Deploying Shredder Locally

First make sure you have the following environment variables defined:
```
AWS_ACCESS_KEY_ID           # Access Key for the aws account that should be shredded
AWS_SECRET_ACCESS_KEY       # Secret Key for the aws account that should be shredded
AWS_ACCOUNTS_TO_SHRED_FILE  # AWS account IDs that should be shredded, 1 id per line
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

Generally speaking, you should first try to shred an account by finding the official Account CR on the appropriate hive cluster and setting its state to "Failed":
```
$ oc get accounts -n aws-account-operator -o json | jq -r '.items[] | select(.spec.awsAccountID=="AWS_ACC_ID_1234") | "\(.metadata.name)"'
account-cr-name-for-AWS_ACC_ID_1234
$ osdctl account set account-cr-name-for-AWS_ACC_ID_1234 --state=Failed
```

If you cant do this for some reason, you can deploy the AWS Account Shredder locally, create an Account CR for the AWS Account IDs, mark them failed and let your
local shredder clean them up. Use cases for this are predominately around cleaning up orphaned accounts from developer activity in staging/integration environments 
(the shredder should not be used for customer accounts in production). In other words, this method should only be used as a last resort for AWS resources
with no associated Account CR or hive cluster.

> DOUBLE CHECK THAT THE ACCOUNT IDS IN THE `AWS_ACCOUNTS_TO_SHRED_FILE` FILE BEFORE PROCEEDING, AS THIS IS A DESTRUCTIVE OPERATION THAT CAN NOT BE UNDONE!

After deploying the AWS Account Shredder locally and setting `AWS_ACCOUNTS_TO_SHRED_FILE` run:
```
make shred-accounts
```

Shredding can take minutes per account and sometimes it can take several passes to cleanup all resources, so a large list of AWS Account IDs can take a long time to finish. 
You can check on the status of the shredding with:
```
make shred-accounts-status
```

You can remove successfully shredded accounts to reduce unnecessary shredder work on large AWS Account ID lists with:
```
make shred-accounts-clean
```

Example:
```
$ oc status
In project default on server https://api.crc.testing:6443

svc/openshift - kubernetes.default.svc.cluster.local
svc/kubernetes - 10.217.4.1:443 -> 6443

View details with 'oc describe <resource>/<name>' or list resources with 'oc get all'.
$ oc whoami
kubeadmin
$ cat ~/aws_account_ids.txt
1234
9876
$ export AWS_ACCOUNTS_TO_SHRED_FILE=~/aws_account_ids.txt
$ make predeploy
...
$ make deploy
...
$ make shred-accounts
hack/get_current_api_url.sh | grep '127.0.0.1\|api.crc.testing'
https://api.crc.testing:6443
hack/shred_accounts.sh -f /Users/mstratto/aws_account_ids.txt mark

Marking accounts for shredding

account.aws.managed.openshift.io/aws-shredder-account-delete-1234 created
account.aws.managed.openshift.io/aws-shredder-account-delete-9876 created
$ make shred-accounts-status
hack/get_current_api_url.sh | grep '127.0.0.1\|api.crc.testing'
https://api.crc.testing:6443
hack/shred_accounts.sh -f /Users/mstratto/aws_account_ids.txt status

Checking account shredder status.

1234 - pending
9876 - Ready

Accounts: 2

Accounts shredded: 1
Account pending: 1
Accounts missing: 0
$ make shred-accounts-cleanup
hack/get_current_api_url.sh | grep '127.0.0.1\|api.crc.testing'
https://api.crc.testing:6443
hack/shred_accounts.sh -f /Users/mstratto/aws_account_ids.txt cleanup

Cleaning up after account shredder.

1234 - pending
9876 - Ready
account.aws.managed.openshift.io "aws-shredder-account-delete-9876" deleted

Accounts: 2

Accounts shredded: 1
Account pending: 1
Accounts missing: 0
```

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
