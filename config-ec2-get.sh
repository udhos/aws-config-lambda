#!/bin/bash

me=$(basename $0)

msg() {
	echo >&2 $me: $@
}

cleanup() {
	[ -f "$tmp" ] && rm "$tmp"
}

die() {
	msg $@
	cleanup
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
.configurationItemMD5Hash
.arn
.configurationItemCaptureTime
.accountId
.configurationStateId
.relationships[].relationshipName
.kernelId
__EOF__
}

exclude_config() {
	cat <<__EOF__
.networkInterfaces[].interfaceType
__EOF__
}

filter() {
	# extract only first item
	local exc=$(exclude | paste -s -d ,)
	jq -r '.configurationItems[0]' | jq -r "del($exc)"
}

# remove fields from .configuration (json encoded as string)
filter_config() {
	local orig=$(mktemp -t filter_config_orig.XXXXXXXXXX)
	cat >$orig
	local exc=$(exclude_config | paste -s -d ,)
	local t=$(mktemp -t filter_config_tmp.XXXXXXXXXX)
	jq -r ".configuration | fromjson | del($exc) | tostring" < $orig > $t
	jq -r ".configuration = \"$(sed -e 's/"/\\\"/g' < $t)\"" < $orig
	rm $orig $t
}

tmp=$resource_id.tmp

aws configservice get-resource-config-history --max-items 1 --resource-type AWS::EC2::Instance --resource-id $resource_id > $tmp || die failure fetching resource

filter < $tmp | filter_config > $resource_id

cleanup

msg saved resource as: $resource_id
