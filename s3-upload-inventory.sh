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

[ -f $resource_id ] || die "missing file: resource-id=[$resource_id]"

aws s3 cp $resource_id.inventory s3://$bucket/$resource_id

