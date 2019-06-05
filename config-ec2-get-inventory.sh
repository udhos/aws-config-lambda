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

exclude() {
	cat <<__EOF__
.version
.arn
.configurationItemCaptureTime
.accountId
.relationships[].relationshipName
__EOF__
}

filter() {
	# extract only first item
	local exc=$(exclude | paste -s -d ,)
	jq -r '.configurationItems[0]' | jq -r "del($exc)"
}

out=$resource_id.inventory

aws configservice get-resource-config-history --max-items 1 --resource-type AWS::SSM::ManagedInstanceInventory --resource-id $resource_id | filter > $out || die failure fetching resource

msg saved resource as: $out

