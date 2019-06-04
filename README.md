# aws-config-lambda

aws-config-lambda: Lambda function in Go for detecting configuration drift in AWS Config Rules.

## Scripts

Build and deploy lambda function:

    ./0-build.sh                      ;# build lambda function
    ./1-create.sh                     ;# create lambda function on aws
    ./2-update.sh                     ;# upload new lambda function to aws

Save resources' state:

    ./run-config-save-all.sh bucket   ;# upload all resources' config to s3 bucket

Helper scripts:

    ./config-ec2-get.sh resource-id   ;# download resource config
    ./ec2-list-by-tag.sh              ;# list resources by tag
    ./s3-upload.sh resource-id bucket ;# upload single resource config to s3

## Rule parameters

Parameters for AWS Config Rules.

- Bucket: Required. Bucket storing desired configurations.

- Dump: Optional. If defined as 'ConfigItem', enables verbose logging.

- ResourceTypes: Optional. List of accepted resource types. If defined, restricts allowed resource types. Example value: 'AWS::EC2::Instance'. You can use 'AWS::SSM::ManagedInstanceInventory' to handle Systems Manager Inventory recorded as AWS Config configuration item.

- TopicArn: Optional. If defined, will publish non-compliance alerts. Example value: arn:aws:sns:sa-east-1:0123456789012:topic-name-for-non-compliance

- ForceNonCompliance: Optional. If defined, evaluations will report non-compliance.
