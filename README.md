# aws-config-lambda

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

- Bucket: Required.
- Dump: Optional. If defined as 'ConfigItem', enables verbose logging.
- ResourceTypes: Optional. List of accepted resource types. If defined, restricts allowed resource types. Example value: 'AWS::EC2::Instance'
- TopicArn: Optional. Example value: arn:aws:sns:sa-east-1:0123456789012:bucket-name-non-compliance
- ForceNonCompliance: Optional. If defined, evaluations will report non-compliance.
