# aws-account-shredder
Repository to audit, service, and clean up leftover AWS resources

## Deploying Shredder Locally

Load up CRC or Minishift. If you don't already have them, create the namespace for `aws-account-operator` and `aws-account-shredder`.

Apply the following secret, filling in your own account details based on the environment you're creating:

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

Open the aws-account-shredder repository and apply the `service_account.yaml`, `service_account_role.yaml`, and `service_account_rolebinding.yaml` to the `aws-account-shredder` namespace.  Apply the `read_account_role.yaml`, `read_account_role_binding.yaml` files to the `aws-account-operator` namespace.

Assert that you have no failed accounts in the `aws-account-operator` namespace, otherwise these will be shredded once you run the next step.

Apply the `deployment.yaml` file.

## Running an ad-hoc shred

The following steps are for running a shred on a single account.

This is best done without the aws-account-operator running.

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

Once you are done with the cleanup, remove the Failed account (otherwise the shredder will infinitely loop over this account).  You can accomplish this with `oc delete -n aws-account-operator aws-shredder-account-delete`
