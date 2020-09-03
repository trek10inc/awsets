package resource

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"

	context2 "github.com/trek10inc/awsets/context"
)

func Test_NewResourceWithTags(t *testing.T) {
	config := aws.Config{
		Region: "us-east-1",
	}
	ctx := context2.AWSetsCtx{
		AWSCfg:    config,
		AccountId: "123456789",
		Context:   context.Background(),
		Logger:    nil,
	}
	object := map[string]interface{}{
		"Foo": "Bar",
		"Tags": map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		},
	}
	r := New(ctx, Ec2Instance, "resource_id", "resource_name", object)
	if r.Tags["tag1"] != "value1" {
		t.Fatalf("expected tag that was not present\n")
	}
	if r.Id != "resource_id" {
		t.Fatalf("expected %s, got %s\n", "resource_id", r.Id)
	}
	if r.Id != "resource_name" {
		t.Fatalf("expected %s, got %s\n", "resource_name", r.Name)
	}

}

func Test_toString(t *testing.T) {
	a := "test"
	if a != toString(a) {
		t.Fatalf("expected %s, got %v\n", a, toString(a))
	}
	if a != toString(&a) {
		t.Fatalf("expected %s, got %v\n", a, toString(a))
	}
	if "" != toString(nil) {
		t.Fatalf("expected %s, got %v\n", "\"\"", toString(a))
	}
	var b *string
	if "" != toString(b) {
		t.Fatalf("expected %s, got %v\n", "\"\"", toString(b))
	}

	c := 1
	if "" != toString(&c) {
		t.Fatalf("expected %s, got %v\n", "\"\"", toString(c))
	}
}
