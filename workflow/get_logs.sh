#!/bin/bash
POD_NAME=$(oc get pod -n aws-account-shredder -ojson | jq -r '.items[]| (.metadata.name)')
oc -n aws-account-shredder logs -f pod/$POD_NAME