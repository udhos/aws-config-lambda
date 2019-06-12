#!/bin/bash

me=$(basename $0)

msg() {
	echo >&2 $me: $@
}

die() {
	msg $@
	exit 1
}

if [ $# -lt 2 ]; then
	echo >&2 usage: $me resource-id bucket
	exit 1
fi

resource_id=$1
bucket=$2

resource_file="$resource_id.inventory"

[ -f $resource_file ] || die "missing resource file: [$resource_file]"

if echo $bucket | grep -q /$; then
	cmd="aws s3 cp $resource_file s3://$bucket$resource_id"
else
	cmd="aws s3 cp $resource_file s3://$bucket/$resource_id"
fi

echo $cmd

$cmd

