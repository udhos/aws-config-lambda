#!/bin/bash

msg() {
	echo >&2 $0: $@
}

die() {
	msg $@
	exit 1
}

filters='Name=tag:group,Values=ssm-lab'
[ -z "$FILTERS" ] || filters="$FILTERS"

msg FILTERS=[$FILTERS] filters=[$filters]

[[ -z "${filters// }" ]] && die "refusing to run with empty filters=[$filters]"

aws ec2 describe-instances --filters "$filters" | jq -r '.Reservations[].Instances[].InstanceId'

