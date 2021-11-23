apiVersion: aws.managed.openshift.io/v1alpha1
kind: Account
metadata:
  name: ${SHRED_ACCOUNT_NAME}
  namespace: ${NAMESPACE}
spec:
  awsAccountID: "${SHRED_ACCOUNT_ID}"
  claimLink: ""
  iamUserSecret: ""
  legalEntity:
    id: ""
    name: ""
