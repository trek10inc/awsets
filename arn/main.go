package arn

import (
	"strings"
)

type Arn struct {
	Raw             string
	ARN             string
	Partition       string
	Service         string
	Region          string
	Account         string
	ResourceType    string
	ResourceId      string
	ResourceVersion string
}

func IsArnP(arn *string) bool {
	if arn == nil {
		return false
	}
	return IsArn(*arn)
}

func IsArn(arn string) bool {
	//TODO make more sophisticated
	return strings.HasPrefix(arn, "arn:")
}

func ParseP(arn *string) Arn {
	return Parse(*arn)
}

func injectResourceIdAndVersion(arn Arn, resourceArnPart string) Arn {
	resourceParts := strings.SplitN(resourceArnPart, ":", 2)
	arn.ResourceId = resourceParts[0]
	if len(resourceParts) == 2 {
		arn.ResourceVersion = resourceParts[1]
	}
	return arn
}

func Parse(arn string) Arn {
	split := strings.SplitN(arn, ":", 6)
	ret := Arn{
		Raw:       arn,
		ARN:       split[0],
		Partition: split[1],
		Service:   split[2],
		Region:    split[3],
		Account:   split[4],
	}

	colonPos := strings.Index(split[5], ":")
	slashPos := strings.Index(split[5], "/")

	// type of resource followed by a colon (:)
	// example: arn:aws:logs:us-east-1:111000111000:log-group:/aws/kinesisfirehose/aws-waf-logs-us-east-1-analytics-us2:*
	if colonPos != -1 && (slashPos == -1 || slashPos > colonPos) {
		resourceParts := strings.SplitN(split[5], ":", 2)
		ret.ResourceType = resourceParts[0]
		return injectResourceIdAndVersion(ret, resourceParts[1])
	}

	// no resource type
	// example: arn:aws:sns:eu-west-2:111000111000:foo
	if m := strings.Count(split[5], "/"); m == 0 {
		return injectResourceIdAndVersion(ret, split[5])
	}

	// type of resource followed by a slash (/)
	// example: arn:aws:ecs:us-east-2:111000111000:task-definition/test-OHKJUeQfdbdv:1
	resourceParts := strings.SplitN(split[5], "/", 2)
	ret.ResourceType = resourceParts[0]
	return injectResourceIdAndVersion(ret, resourceParts[1])
}
