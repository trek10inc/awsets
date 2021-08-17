module github.com/trek10inc/awsets

go 1.17

require (
	github.com/aws/aws-sdk-go-v2 v1.8.0
	github.com/aws/aws-sdk-go-v2/config v1.6.0
	github.com/aws/aws-sdk-go-v2/service/accessanalyzer v1.5.2
	github.com/aws/aws-sdk-go-v2/service/acm v1.5.1
	github.com/aws/aws-sdk-go-v2/service/amplify v1.4.1
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.5.2
	github.com/aws/aws-sdk-go-v2/service/apigatewayv2 v1.4.2
	github.com/aws/aws-sdk-go-v2/service/applicationautoscaling v1.4.2
	github.com/aws/aws-sdk-go-v2/service/appmesh v1.4.2
	github.com/aws/aws-sdk-go-v2/service/appsync v1.5.0
	github.com/aws/aws-sdk-go-v2/service/athena v1.5.0
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.11.0
	github.com/aws/aws-sdk-go-v2/service/backup v1.4.2
	github.com/aws/aws-sdk-go-v2/service/batch v1.6.0
	github.com/aws/aws-sdk-go-v2/service/budgets v1.3.2
	github.com/aws/aws-sdk-go-v2/service/cloud9 v1.5.2
	github.com/aws/aws-sdk-go-v2/service/cloudformation v1.8.0
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.7.1
	github.com/aws/aws-sdk-go-v2/service/cloudtrail v1.4.2
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.7.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatchevents v1.5.2
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.5.2
	github.com/aws/aws-sdk-go-v2/service/codebuild v1.7.0
	github.com/aws/aws-sdk-go-v2/service/codecommit v1.3.2
	github.com/aws/aws-sdk-go-v2/service/codedeploy v1.5.2
	github.com/aws/aws-sdk-go-v2/service/codepipeline v1.4.2
	github.com/aws/aws-sdk-go-v2/service/codestar v1.3.2
	github.com/aws/aws-sdk-go-v2/service/cognitoidentity v1.4.2
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.5.0
	github.com/aws/aws-sdk-go-v2/service/configservice v1.6.2
	github.com/aws/aws-sdk-go-v2/service/databasemigrationservice v1.6.1
	github.com/aws/aws-sdk-go-v2/service/dax v1.3.2
	github.com/aws/aws-sdk-go-v2/service/docdb v1.8.1
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.4.2
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.3.2
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.13.0
	github.com/aws/aws-sdk-go-v2/service/ecr v1.4.2
	github.com/aws/aws-sdk-go-v2/service/ecs v1.8.0
	github.com/aws/aws-sdk-go-v2/service/efs v1.5.2
	github.com/aws/aws-sdk-go-v2/service/eks v1.8.1
	github.com/aws/aws-sdk-go-v2/service/elasticache v1.8.1
	github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk v1.5.1
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.5.1
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.6.0
	github.com/aws/aws-sdk-go-v2/service/elasticsearchservice v1.5.2
	github.com/aws/aws-sdk-go-v2/service/emr v1.5.0
	github.com/aws/aws-sdk-go-v2/service/firehose v1.4.2
	github.com/aws/aws-sdk-go-v2/service/fsx v1.7.2
	github.com/aws/aws-sdk-go-v2/service/glue v1.10.0
	github.com/aws/aws-sdk-go-v2/service/greengrass v1.5.2
	github.com/aws/aws-sdk-go-v2/service/guardduty v1.4.2
	github.com/aws/aws-sdk-go-v2/service/iam v1.8.0
	github.com/aws/aws-sdk-go-v2/service/imagebuilder v1.8.0
	github.com/aws/aws-sdk-go-v2/service/iot v1.9.0
	github.com/aws/aws-sdk-go-v2/service/iotsitewise v1.9.0
	github.com/aws/aws-sdk-go-v2/service/kafka v1.5.2
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.5.2
	github.com/aws/aws-sdk-go-v2/service/kms v1.4.2
	github.com/aws/aws-sdk-go-v2/service/lambda v1.6.0
	github.com/aws/aws-sdk-go-v2/service/mq v1.4.1
	github.com/aws/aws-sdk-go-v2/service/neptune v1.7.1
	github.com/aws/aws-sdk-go-v2/service/qldb v1.5.0
	github.com/aws/aws-sdk-go-v2/service/rds v1.7.0
	github.com/aws/aws-sdk-go-v2/service/redshift v1.10.0
	github.com/aws/aws-sdk-go-v2/service/route53 v1.9.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.12.0
	github.com/aws/aws-sdk-go-v2/service/sagemaker v1.11.0
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.5.0
	github.com/aws/aws-sdk-go-v2/service/servicecatalog v1.4.2
	github.com/aws/aws-sdk-go-v2/service/servicediscovery v1.7.2
	github.com/aws/aws-sdk-go-v2/service/ses v1.5.1
	github.com/aws/aws-sdk-go-v2/service/sfn v1.4.2
	github.com/aws/aws-sdk-go-v2/service/signer v1.4.2
	github.com/aws/aws-sdk-go-v2/service/sns v1.7.1
	github.com/aws/aws-sdk-go-v2/service/sqs v1.7.1
	github.com/aws/aws-sdk-go-v2/service/ssm v1.9.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.6.1
	//github.com/aws/aws-sdk-go-v2/service/timestreamquery v1.1.1
	//github.com/aws/aws-sdk-go-v2/service/timestreamwrite v1.1.1
	github.com/aws/aws-sdk-go-v2/service/transfer v1.5.2
	github.com/aws/aws-sdk-go-v2/service/waf v1.3.2
	github.com/aws/aws-sdk-go-v2/service/wafregional v1.4.2
	github.com/aws/aws-sdk-go-v2/service/wafv2 v1.7.0
	github.com/aws/aws-sdk-go-v2/service/workspaces v1.5.2
	github.com/fatih/structs v1.1.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.3.2 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.4.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.0.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.2.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.0.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.2.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.5.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.3.2 // indirect
	github.com/aws/smithy-go v1.7.0 // indirect
)
