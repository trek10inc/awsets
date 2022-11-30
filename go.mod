module github.com/trek10inc/awsets

go 1.17

require (
	github.com/aws/aws-sdk-go-v2 v1.16.4
	github.com/aws/aws-sdk-go-v2/config v1.15.7
	github.com/aws/aws-sdk-go-v2/service/accessanalyzer v1.15.5
	github.com/aws/aws-sdk-go-v2/service/acm v1.14.5
	github.com/aws/aws-sdk-go-v2/service/amplify v1.11.6
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.15.5
	github.com/aws/aws-sdk-go-v2/service/apigatewayv2 v1.12.5
	github.com/aws/aws-sdk-go-v2/service/applicationautoscaling v1.15.5
	github.com/aws/aws-sdk-go-v2/service/appmesh v1.14.0
	github.com/aws/aws-sdk-go-v2/service/appsync v1.14.5
	github.com/aws/aws-sdk-go-v2/service/athena v1.15.2
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.23.2
	github.com/aws/aws-sdk-go-v2/service/backup v1.16.1
	github.com/aws/aws-sdk-go-v2/service/batch v1.18.3
	github.com/aws/aws-sdk-go-v2/service/budgets v1.12.5
	github.com/aws/aws-sdk-go-v2/service/cloud9 v1.16.5
	github.com/aws/aws-sdk-go-v2/service/cloudformation v1.20.5
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.18.1
	github.com/aws/aws-sdk-go-v2/service/cloudtrail v1.16.1
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.18.3
	github.com/aws/aws-sdk-go-v2/service/cloudwatchevents v1.14.6
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.15.6
	github.com/aws/aws-sdk-go-v2/service/codebuild v1.19.5
	github.com/aws/aws-sdk-go-v2/service/codecommit v1.13.5
	github.com/aws/aws-sdk-go-v2/service/codedeploy v1.14.5
	github.com/aws/aws-sdk-go-v2/service/codepipeline v1.13.5
	github.com/aws/aws-sdk-go-v2/service/codestar v1.11.5
	github.com/aws/aws-sdk-go-v2/service/cognitoidentity v1.13.5
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.15.5
	github.com/aws/aws-sdk-go-v2/service/configservice v1.21.2
	github.com/aws/aws-sdk-go-v2/service/databasemigrationservice v1.18.5
	github.com/aws/aws-sdk-go-v2/service/dax v1.11.5
	github.com/aws/aws-sdk-go-v2/service/docdb v1.18.2
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.15.5
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.13.5
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.43.1
	github.com/aws/aws-sdk-go-v2/service/ecr v1.17.5
	github.com/aws/aws-sdk-go-v2/service/ecs v1.18.7
	github.com/aws/aws-sdk-go-v2/service/efs v1.17.3
	github.com/aws/aws-sdk-go-v2/service/eks v1.21.1
	github.com/aws/aws-sdk-go-v2/service/elasticache v1.20.7
	github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk v1.14.5
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.14.5
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.18.5
	github.com/aws/aws-sdk-go-v2/service/elasticsearchservice v1.15.5
	github.com/aws/aws-sdk-go-v2/service/emr v1.18.1
	github.com/aws/aws-sdk-go-v2/service/firehose v1.14.6
	github.com/aws/aws-sdk-go-v2/service/fsx v1.23.2
	github.com/aws/aws-sdk-go-v2/service/glue v1.25.0
	github.com/aws/aws-sdk-go-v2/service/greengrass v1.13.5
	github.com/aws/aws-sdk-go-v2/service/guardduty v1.13.6
	github.com/aws/aws-sdk-go-v2/service/iam v1.18.5
	github.com/aws/aws-sdk-go-v2/service/imagebuilder v1.19.5
	github.com/aws/aws-sdk-go-v2/service/iot v1.25.2
	github.com/aws/aws-sdk-go-v2/service/iotsitewise v1.21.2
	github.com/aws/aws-sdk-go-v2/service/kafka v1.17.5
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.15.6
	github.com/aws/aws-sdk-go-v2/service/kms v1.17.2
	github.com/aws/aws-sdk-go-v2/service/lambda v1.23.1
	github.com/aws/aws-sdk-go-v2/service/mq v1.13.1
	github.com/aws/aws-sdk-go-v2/service/neptune v1.16.5
	github.com/aws/aws-sdk-go-v2/service/qldb v1.14.6
	github.com/aws/aws-sdk-go-v2/service/rds v1.21.2
	github.com/aws/aws-sdk-go-v2/service/redshift v1.24.1
	github.com/aws/aws-sdk-go-v2/service/route53 v1.20.5
	github.com/aws/aws-sdk-go-v2/service/s3 v1.26.10
	github.com/aws/aws-sdk-go-v2/service/sagemaker v1.30.1
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.15.8
	github.com/aws/aws-sdk-go-v2/service/servicecatalog v1.14.3
	github.com/aws/aws-sdk-go-v2/service/servicediscovery v1.17.5
	github.com/aws/aws-sdk-go-v2/service/ses v1.14.5
	github.com/aws/aws-sdk-go-v2/service/sfn v1.13.5
	github.com/aws/aws-sdk-go-v2/service/signer v1.13.5
	github.com/aws/aws-sdk-go-v2/service/sns v1.17.6
	github.com/aws/aws-sdk-go-v2/service/sqs v1.18.5
	github.com/aws/aws-sdk-go-v2/service/ssm v1.27.1
	github.com/aws/aws-sdk-go-v2/service/sts v1.16.6
	//github.com/aws/aws-sdk-go-v2/service/timestreamquery v1.1.1
	//github.com/aws/aws-sdk-go-v2/service/timestreamwrite v1.1.1
	github.com/aws/aws-sdk-go-v2/service/transfer v1.19.0
	github.com/aws/aws-sdk-go-v2/service/waf v1.11.5
	github.com/aws/aws-sdk-go-v2/service/wafregional v1.12.6
	github.com/aws/aws-sdk-go-v2/service/wafv2 v1.20.1
	github.com/aws/aws-sdk-go-v2/service/workspaces v1.18.2
	github.com/fatih/structs v1.1.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

require gopkg.in/yaml.v3 v3.0.1

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.12.2 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.11 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.12 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.7.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.13.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.5 // indirect
	github.com/aws/smithy-go v1.11.2 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
)
