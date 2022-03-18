#!/bin/bash

ACCOUNT_OPERATOR_NAMESPACE=aws-account-operator
ACCOUNT_SHREDDER_NAMESPACE=aws-account-shredder
OC_LOG_WINDOW="1h"


function help {
  name=`basename $0`
  cat <<EOF
$name [-f filename] [-w ocLogsSince] [-h] command

Run an ad-hoc shredding of 1 or more AWS account ids using a locally deployed AWS Account Shredder.

For more details see: https://github.com/openshift/aws-account-shredder#running-an-ad-hoc-shred

Requirements:
    1. `oc` and `osdctl` available in your $PATH
    2. AWS Account Shredder running locally - https://github.com/openshift/aws-account-shredder#deploying-shredder-locally 

commands:
    mark        marks Account CRs of the associated AWS account ids as failed so the 
                shredder cleans up their AWS resources.
    status      check the shredding status of the AWS account ids.
    cleanup     remove Account CRs of AWS account ids.

options:
    -h  print script usage message and exit
    -f  file path to a list of AWS account ids. This file should only contain
        AWS account ids, 1 per line.
    -w  the "since" variable used when fetching oc logs (e.g. `oc logs --since 5m`).
        Shorter windows will speed up runtimes. Defaults to "1h"
EOF
}

function lowerCase {
    echo "$1" | awk '{print tolower($0)}'
}

function awsAccountIdsFromFile {
    local file=$1
    while IFS=" " read -r AWS_ACC_ID
    do
        echo $AWS_ACC_ID
    done < "$file"
}

function accountCrName {
    local awsAccountId=$1
    echo "aws-shredder-account-delete-$awsAccountId"
}

# generate a json map of AWS Account IDs to oc state which will act as a cache to reduce overall oc calls which are relatively slow
# e.g. {"1234": "Failed", "9876": "Ready", ...}
function ocAccountStatusCache {
    oc get accounts -n $ACCOUNT_OPERATOR_NAMESPACE -o json | jq '.items | map({ (.spec.awsAccountID): .status.state }) | add'
}

# generate a json map of AWS Account IDs that have the "Failed to reset account status" error message in logs that 
# indicates an account has been successfully shredded. oc logs can be quite large and this reduces the ammount of 
# times we need jq to process the large amount of json
# e.g. {"1234": true, "9876": true, ...}
function accountFailedToUpdateCache {
    oc logs deployment/aws-account-shredder -n $ACCOUNT_SHREDDER_NAMESPACE --since $OC_LOG_WINDOW | \
    jq -R "fromjson? | . " -c | \
    jq -s '[.[] | select(.msg | contains("Failed to reset account status"))] | map({(.AccountID): true}) | add'
}

# Reads AWS Account IDs from a file and checks on the shredding status of the Account CR
# can optionally remove Account CRs it determines have been successfully shredded
function checkAccountShredderStatus {    
    accountIds=( $(awsAccountIdsFromFile $1) )
    cleanup=$2
    shredded=0
    pending=0
    missing=0

    echo "Gathering data from oc (this can take a while)..."
    accountStatusJson=$(ocAccountStatusCache)
    accountsFailedToUpdateStatus=$(accountFailedToUpdateCache)
    
    for id in "${accountIds[@]}"
    do
        # lookup oc account state in our cache
        accountCrState=$(echo $accountStatusJson | jq -r --arg id "$id" '.[$id]')
        accountCrName=$(accountCrName $id)

        if [ "$accountCrState" = "" ] || [ "$accountCrState" = "null" ]; then
            echo "$id - Account CR missing"
            missing=$((missing+1))
        elif [ "$accountCrState" = "Failed" ]; then
            #check for "Failed to reset account status" message in our cache
            failedToUpdateStatusLogMessagePresent=$(echo $accountsFailedToUpdateStatus | jq -r --arg id "$id" '.[$id]')
            if [ "$failedToUpdateStatusLogMessagePresent" = "true" ]; then
                echo "$id - shredded (unable to update status)"
                shredded=$((shredded+1))
                if [ "$cleanup" = true ] ; then
                    oc delete account -n $ACCOUNT_OPERATOR_NAMESPACE $accountCrName
                fi
            else
                echo "$id - pending"
                echo "$id" >> pending.txt
                pending=$((pending+1))
            fi
        else
            echo "$id - $accountCrState"
            shredded=$((shredded+1))
            if [ "$cleanup" = true ] ; then
                oc delete account -n $ACCOUNT_OPERATOR_NAMESPACE $accountCrName
            fi
        fi
    done

    total=$((shredded + pending + missing))
    echo ""
    echo "Accounts: $total"
    echo ""
    echo "Accounts shredded: $shredded"
    echo "Account pending: $pending"
    echo "Accounts missing: $missing"
}

function markAccountForShredding {
    echo "Gathering data from oc (this can take a while)..."
    accountStatusJson=$(ocAccountStatusCache)
    accountIds=( $(awsAccountIdsFromFile $1) )

    for id in "${accountIds[@]}"
    do
        accountCrName=$(accountCrName $id)
        accountCrState=$(echo $accountStatusJson | jq -r --arg id "$id" '.[$id]')
        if [ "$accountCrState" = "" ] || [ "$accountCrState" = "null" ]; then
            oc process --local -f hack/templates/aws-shredder-account-delete-template.yaml \
                -p NAMESPACE=$ACCOUNT_OPERATOR_NAMESPACE \
                -p SHRED_ACCOUNT_CR_NAME=$accountCrName \
                -p ACCOUNT_ID_TO_SHRED=$id | \
                oc apply -f -
        else
            echo "Account CR $accountCrName already exists."
        fi
        osdctl account set $accountCrName --state=Failed
    done
}

fileName=""

while getopts "h:f:w:" flag; do
case "$flag" in
    h) help
       exit 0;;
    f) fileName=$OPTARG;;
    w) OC_LOG_WINDOW=$OPTARG;;
esac
done

if [ ! -f "$fileName" ]; then
    echo "ERROR - AWS Account IDs file $fileName does not exist."
    exit 1
fi

cmd=`lowerCase ${@:$OPTIND:1}`

case "$cmd" in
    "status")
        echo ""
        echo "Checking account shredder status."
        checkAccountShredderStatus $fileName false
        ;;
    "cleanup")
        echo ""
        echo "Cleaning up after account shredder."
        checkAccountShredderStatus $fileName true
        ;;
    "mark")
        echo ""
        echo "Marking accounts for shredding"
        markAccountForShredding $fileName
        ;;
    *)
        echo "ERROR - unknown command: $cmd"
        help
        exit 2
        ;;
esac
