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

func Parse(arn string) Arn {
	//"arn:aws:ecs:us-east-1:067142875141:cluster/baseinfr-ECSCluster-1GQBLCJV73B1G"
	split := strings.SplitN(arn, ":", 6)
	ret := Arn{
		Raw:       arn,
		ARN:       split[0],
		Partition: split[1],
		Service:   split[2],
		Region:    split[3],
		Account:   split[4],
		//rest:      split[5],
	}

	if n := strings.Count(split[5], ":"); n > 0 {
		resourceParts := strings.SplitN(split[5], ":", 2)
		ret.ResourceType = resourceParts[0]
		ret.ResourceId = resourceParts[1]
	} else {
		if m := strings.Count(split[5], "/"); m == 0 {
			ret.ResourceId = split[5]
		} else {
			resourceParts := strings.SplitN(split[5], "/", 2)
			ret.ResourceType = resourceParts[0]
			ret.ResourceId = resourceParts[1]
		}
	}
	// log group: "arn:aws:logs:us-east-1:067142875141:log-group:/aws/kinesisfirehose/aws-waf-logs-us-east-1-analytics-us2:*" -> for log group, resource id shouldn't have :* at the end
	return ret
	// TODO: doesn't handle versions
}
