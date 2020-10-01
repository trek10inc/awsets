package arn

import "testing"

func Test_IsArn(t *testing.T) {
	tests := map[string]bool{
		"boop":     false,
		"arn":      false,
		"":         false,
		"arn:boop": true,
	}
	for k, v := range tests {
		t.Run(k, func(t *testing.T) {
			if IsArn(k) != v {
				t.Fail()
			}
		})
	}
}

func stringPointer(input string) *string {
	return &input
}

func Test_IsArnP(t *testing.T) {
	tests := map[*string]bool{
		nil:                       false,
		stringPointer(""):         false,
		stringPointer("arn"):      false,
		stringPointer("arn:boop"): true,
	}
	for k, v := range tests {
		t.Run("", func(t *testing.T) {
			if IsArnP(k) != v {
				t.Fail()
			}
		})
	}
}

func Test_Parse(t *testing.T) {
	tests := map[string]Arn{
		"arn:aws:ecs:us-east-1:111000111000:cluster/test-ECSCluster-3Z3CPPG9GRGKF": Arn{
			Raw:          "arn:aws:ecs:us-east-1:111000111000:cluster/test-ECSCluster-3Z3CPPG9GRGKF",
			ARN:          "arn",
			Partition:    "aws",
			Service:      "ecs",
			Region:       "us-east-1",
			Account:      "111000111000",
			ResourceType: "cluster",
			ResourceId:   "test-ECSCluster-3Z3CPPG9GRGKF",
		},
		"arn:aws:logs:us-east-1:111000111000:log-group:/aws/kinesisfirehose/aws-waf-logs-us-east-1-analytics-us2:*": Arn{
			Raw:             "arn:aws:logs:us-east-1:111000111000:log-group:/aws/kinesisfirehose/aws-waf-logs-us-east-1-analytics-us2:*",
			ARN:             "arn",
			Partition:       "aws",
			Service:         "logs",
			Region:          "us-east-1",
			Account:         "111000111000",
			ResourceType:    "log-group",
			ResourceId:      "/aws/kinesisfirehose/aws-waf-logs-us-east-1-analytics-us2",
			ResourceVersion: "*",
		},
		"arn:aws:sns:eu-west-2:111000111000:foo": Arn{
			Raw:        "arn:aws:sns:eu-west-2:111000111000:foo",
			ARN:        "arn",
			Partition:  "aws",
			Service:    "sns",
			Region:     "eu-west-2",
			Account:    "111000111000",
			ResourceId: "foo",
		},
		"arn:aws:lambda:us-east-1:111000111000:function:foobar": Arn{
			Raw:          "arn:aws:lambda:us-east-1:111000111000:function:foobar",
			ARN:          "arn",
			Partition:    "aws",
			Service:      "lambda",
			Region:       "us-east-1",
			Account:      "111000111000",
			ResourceType: "function",
			ResourceId:   "foobar",
		},
		"arn:aws:lambda:us-east-1:111000111000:function:foobar:$LATEST": Arn{
			Raw:             "arn:aws:lambda:us-east-1:111000111000:function:foobar:$LATEST",
			ARN:             "arn",
			Partition:       "aws",
			Service:         "lambda",
			Region:          "us-east-1",
			Account:         "111000111000",
			ResourceType:    "function",
			ResourceId:      "foobar",
			ResourceVersion: "$LATEST",
		},
		"arn:aws:s3:::test-bucket": Arn{
			Raw:        "arn:aws:s3:::test-bucket",
			ARN:        "arn",
			Partition:  "aws",
			Service:    "s3",
			ResourceId: "test-bucket",
		},
		"arn:aws:ecs:us-east-2:111000111000:task-definition/test-OHKJUeQfdbdv:1": Arn{
			Raw:             "arn:aws:ecs:us-east-2:111000111000:task-definition/test-OHKJUeQfdbdv:1",
			ARN:             "arn",
			Partition:       "aws",
			Service:         "ecs",
			Region:          "us-east-2",
			Account:         "111000111000",
			ResourceType:    "task-definition",
			ResourceId:      "test-OHKJUeQfdbdv",
			ResourceVersion: "1",
		},
	}
	for k, v := range tests {
		t.Run(k, func(t *testing.T) {
			if Parse(k) != v {
				t.Fail()
			}
		})
	}
}
