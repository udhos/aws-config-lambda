#!/bin/bash

die () {
    echo 1>&2 $*
    exit 1
}

region=sa-east-1
role_name=role_config_lambda
lambda_name=FunctionConfigLambda

echo region: $region
echo role: $role_name
echo lambda: $lambda_name

# create role
aws iam create-role --role-name $role_name \
    --assume-role-policy-document '{
    "Statement": [{
      "Effect": "Allow",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }]
  }'

# attach policy
aws iam attach-role-policy --role-name $role_name --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
aws iam attach-role-policy --role-name $role_name --policy-arn arn:aws:iam::aws:policy/service-role/AWSConfigRole
aws iam attach-role-policy --role-name $role_name --policy-arn arn:aws:iam::aws:policy/AmazonSNSFullAccess

# get role arn
ROLE_ARN=`aws iam get-role --role-name $role_name --query 'Role.Arn' --output text`

[ -n "$ROLE_ARN" ] || die "missing env var ROLE_ARN=arn:aws:iam::<account-id>:role/<role>"

echo ROLE_ARN=$ROLE_ARN

# create function
#
# handler: name of the executable binary file within the zip package

aws lambda delete-function --function-name $lambda_name

aws lambda create-function \
    --region $region \
    --function-name $lambda_name \
    --zip-file fileb://./main.zip \
    --runtime go1.x \
    --role "$ROLE_ARN" \
    --handler main

