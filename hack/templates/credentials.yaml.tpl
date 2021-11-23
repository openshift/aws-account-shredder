apiVersion: "v1"
kind: Secret
metadata:
  name: aws-account-shredder-credentials
  namespace: ${NAMESPACE}
data:
  aws_access_key_id: ${AWS_ACCESS_KEY_ID_B64}
  aws_secret_access_key: ${AWS_SECRET_ACCESS_KEY_B64}
type: Opaque
