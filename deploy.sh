PROJECT=ng-docs-slack-notification
BUCKET="deployment-nbaeumer-eu-west-1"
SNS_TOPIC_DISPLAY_NAME="AdministrationNotification"
SLACK_WEBHOOK_URL="AQICAHitEE9ZBAkn4k0Nyux9EfWlAGAryjGrN8cG3sZY4YU0UAHV91iHoWZPThy8gkL8IdTmAAAAizCBiAYJKoZIhvcNAQcGoHsweQIBADB0BgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDOesTWq+3xL+03lW/AIBEIBH+rGs1b74Madi29kG8FO6IBZs3VqGQ+m173BX8+15xrbS7KI+hlX9RxKkhsCIAMBkvy9E+O1GqK0GIuWztIvklEe+HCNa8Q8="
LOG_GROUP="ng-docs"
KMS_KEY_ID="ce17fa91-7b37-4c2c-a98a-d939cf48370f"

# make a build directory to store artifacts
rm -rf build
mkdir build

# make the deployment bucket in case it doesn´t exist
if aws s3 ls s3://$BUCKET 2>&1 | grep -q 'NoSuchBucket'
then
 aws s3 mb s3://$BUCKET
fi

# generate next stage yaml file
aws cloudformation package \
    --template-file template.yaml \
    --output-template-file build/output.yaml \
    --s3-bucket $BUCKET

# the actual deployment step
aws cloudformation deploy \
    --template-file build/output.yaml \
    --parameter-override SnsTopicDisplayName=$SNS_TOPIC_DISPLAY_NAME SlackHookURL=$SLACK_WEBHOOK_URL LogGroup=$LOG_GROUP KmsKeyId=$KMS_KEY_ID\
    --stack-name $PROJECT \
    --capabilities CAPABILITY_NAMED_IAM