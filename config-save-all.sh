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
	echo >&2 usage: $me bucket
	exit 1
fi

[ -x ./ec2-list-by-tag.sh ] || die missing script: ./ec2-list-by-tag.sh
[ -x ./config-ec2-get.sh ] || die missing script: ./config-ec2-get.sh
[ -x ./s3-upload.sh ] || die missing script: ./s3-upload.sh

bucket=$1

./ec2-list-by-tag.sh | while read i; do
	msg resource_id: $i
	./config-ec2-get.sh $i
	echo ./s3-upload.sh $i $bucket
done

