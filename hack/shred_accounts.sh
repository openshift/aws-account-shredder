#!/bin/bash

ACCOUNT_OPERATOR_NAMESPACE=aws-account-operator
OC_LOGS_CACHE_FILE=shredder_status.log


function help {
  name=`basename $0`
  cat <<EOF
$name [-f filename] [-h] command

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
EOF
}

function onExit {
    rm $OC_LOGS_CACHE_FILE 2>/dev/null
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

function accountCrExists {
    local accountCrName=$1
    oc get accounts -n $ACCOUNT_OPERATOR_NAMESPACE $accountCrName -o json &>/dev/null
    status=$?
    if (( $status == 0 )); then
        echo true
    else 
        echo false
    fi
}

# Reads AWS Account IDs from a file and checks on the shredding status of the Account CR
# can optionally remove Account CRs it determines have been successfully shredded
function checkAccountShredderStatus {    
    accountIds=( $(awsAccountIdsFromFile $1) )
    cleanup=$2
    shredded=0
    pending=0
    missing=0

    # cache shredder logs to a file (lines which arent parseable as json)
    oc logs deployment/aws-account-shredder -n aws-account-shredder --since 12h | jq -R "fromjson? | . " -c > $OC_LOGS_CACHE_FILE
    
    for id in "${accountIds[@]}"
    do
        accountCrName=$(accountCrName $id)
        accountExists=$(accountCrExists $accountCrName)
        if [ "$accountExists" = true ]; then
            accountCrState=$(oc get accounts -n $ACCOUNT_OPERATOR_NAMESPACE $accountCrName -o json 2>/dev/null | jq -r .status.state)
            if [ "$accountCrState" = "Failed" ]; then
                #account for the bug where an account was shredded but for some reason kubernettes doesnt let the shredded reset the account status
                failedToUpdateStatus=$(cat shredder_status.log | jq -s --arg id "$id" '.[] | select(.AccountID | . and . == $id) | select(.msg | contains("Failed to reset account status"))' | jq -s '. | length')
                if (( $failedToUpdateStatus > 0 )) ; then
                    echo "$id - shredded (unable to update status)"
                    shredded=$((shredded+1))
                    if [ "$cleanup" = true ] ; then
                        oc delete account -n $ACCOUNT_OPERATOR_NAMESPACE $accountCrName
                    fi
                else
                    echo "$id - pending"
                    pending=$((pending+1))
                fi
            else
                echo "$id - $accountCrState"
                shredded=$((shredded+1))
                if [ "$cleanup" = true ] ; then
                    oc delete account -n $ACCOUNT_OPERATOR_NAMESPACE $accountCrName
                fi
            fi
        else
            echo "$id - Account CR missing"
            missing=$((missing+1))
        fi
    done

    total=$((shredded + pending + missing))
    echo "Accounts: $total"
    echo "Accounts shredded: $shredded"
    echo "Account pending: $pending"
    echo "Accounts missing: $missing"
}

function markAccountForShredding {
    accountIds=( $(awsAccountIdsFromFile $1) )
    for id in "${accountIds[@]}"
    do
        accountCrName=$(accountCrName $id)
        accountExists=$(accountCrExists $accountCrName)
        if [ "$accountExists" = false ]; then
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

trap onExit EXIT
fileName=""

while getopts "h:f:" flag; do
case "$flag" in
    h) help
       exit 0;;
    f) fileName=$OPTARG;;
esac
done

if [ ! -f "$fileName" ]; then
    echo "ERROR - AWS Account IDs file $fileName does not exist."
    exit 1
fi

cmd=`lowerCase ${@:$OPTIND:1}`

case "$cmd" in
    "status")
        echo "Checking account shredder status."
        checkAccountShredderStatus $fileName false
        ;;
    "cleanup")
        echo "Cleaning up after account shredder."
        checkAccountShredderStatus $fileName true
        ;;
    "mark")
        echo "Marking accounts for shredding"
        markAccountForShredding $fileName
        ;;
    *)
        echo "ERROR - unknown command: $cmd"
        help
        exit 2
        ;;
esac
