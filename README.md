# aws-account-shredder
[![Go Report Card](https://goreportcard.com/badge/github.com/openshift/aws-account-shredder)](https://goreportcard.com/report/github.com/openshift/aws-account-shredder)
[![GoDoc](https://godoc.org/github.com/openshift/aws-account-shredder?status.svg)](https://pkg.go.dev/mod/github.com/openshift/aws-account-shredder)
[![License](https://img.shields.io/:license-apache-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)


Repository to audit, service, and clean up leftover AWS resources

## Deploying Shredder Locally

Load up CRC or Minishift. If you don't already have them, create the namespaces:
```
oc create ns aws-account-operator
oc create ns aws-account-shredder
```

You'll need to apply a secret with your aws account details, the details for your aws credentials will need to be base64 encoded before being added to the secret, you can do that like:
```
echo -n {AWS_ACCESS_KEY_ID} | base64
echo -n {AWS_SECRET_ACCESS_KEY} | base64
```

Now you can fill in the fields below and apply the secret:
```json
{
    "apiVersion": "v1",
    "data": {
        "aws_access_key_id": "",
        "aws_secret_access_key": ""
},
    "kind": "Secret",
    "metadata": {
        "name": "aws-account-shredder-credentials",
        "namespace": "aws-account-shredder"
    },
    "type": "Opaque"
}
```
> **Note:** If you find yourself changing the secret while the shredder is running, you'll need to kill the shredder pod to ensure the updated secret is pulled when the new pod is created.

Open the aws-account-shredder repository and apply the `service_account.yaml`, `service_account_role.yaml`, and `service_account_rolebinding.yaml` to the `aws-account-shredder` namespace.  Apply the `read_account_role.yaml`, `read_account_role_binding.yaml` files to the `aws-account-operator` namespace.

Assert that you have no failed accounts in the `aws-account-operator` namespace, otherwise these will be shredded once you run the next step.

Apply the `deployment.yaml` file.

## Running an ad-hoc shred

The following steps are for running a shred on a single account.

This is best done without the aws-account-operator running.

In order to create an account CR, you'll need to have the CRD defined:
```
cd deploy && curl -O  https://raw.githubusercontent.com/openshift/aws-account-operator/master/deploy/crds/aws.managed.openshift.io_accounts.yaml
oc apply -f aws.managed.openshift.io_accounts.yaml
```

Apply the following JSON, changing the Account ID appropriately.  DOUBLE CHECK THAT YOU ARE ADDING THE CORRECT ACCOUNT ID, AS THIS IS A DESTRUCTIVE OPERATION AND YOU CANNOT UNDO

```json
{
  "apiVersion": "aws.managed.openshift.io/v1alpha1",
  "kind": "Account",
  "metadata": {
    "name": "aws-shredder-account-delete",
    "namespace": "aws-account-operator"
  },
  "spec": {
    "awsAccountID": "",
    "claimLink": "",
    "iamUserSecret": "",
    "legalEntity": {
      "id": "",
      "name": ""
    }
  }
}
```

Using [osdctl](https://github.com/openshift/osd-utils-cli), set the Account State to be Failed:

```bash
osdctl account set aws-shredder-account-delete --state=Failed
```

Once you set the status of the account to failed, the Shredder should pick it up and start shredding through the accounts.

You should be able to follow the logs and watch the shred happen using `oc logs -f [pod name] -n aws-account-shredder`.  Certain objects may not delete on the first attempt through the shredder, but the shredder will continue to run on the account until it is created.

Once you are done with the cleanup, remove the Failed account (otherwise the shredder will infinitely loop over this account).  You can accomplish this with `oc delete -n account aws-account-operator aws-shredder-account-delete`

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