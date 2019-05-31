#!/bin/bash

me=$(basename $0)

msg() {
	echo >&2 $me: $@
}

die() {
	msg $@
	exit 1
}

if [ $# -lt 1 ]; then
	echo >&2 usage: $me resource-id
	exit 1
fi

hash jq || die missing jq

resource_id=$1

aws configservice get-resource-config-history --max-items 1 --resource-type AWS::EC2::Instance --resource-id $resource_id | jq -r '.configurationItems[0]' > $resource_id || die failure fetching resource

msg saved resource as: $resource_id
