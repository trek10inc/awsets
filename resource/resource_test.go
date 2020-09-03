package resource

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"

	context2 "github.com/trek10inc/awsets/context"
)

func Test_NewResourceWithTags(t *testing.T) {

	ctx := getContext()

	object := map[string]interface{}{
		"Foo": "Bar",
		"Tags": map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		},
	}
	r := New(ctx, Ec2Instance, "resource_id", "resource_name", object)
	if r.Region != "us-east-1" {
		t.Fatalf("expected us-east-1, got %s\n", r.Region)
	}
	if r.Tags["tag1"] != "value1" {
		t.Fatalf("expected tag that was not present\n")
	}
	if r.Id != "resource_id" {
		t.Fatalf("expected %s, got %s\n", "resource_id", r.Id)
	}
	if r.Name != "resource_name" {
		t.Fatalf("expected %s, got %s\n", "resource_name", r.Name)
	}
	if r.Version != "" {
		t.Fatalf("expected empty version, got %s\n", r.Version)
	}
}

func Test_NewResourceWithoutTags(t *testing.T) {

	ctx := getContext()

	object := map[string]interface{}{
		"Foo": "Bar",
	}
	r := New(ctx, Ec2Instance, "resource_id", "resource_name", object)
	if len(r.Tags) != 0 {
		t.Fatalf("expected zero tags\n")
	}
	if r.Id != "resource_id" {
		t.Fatalf("expected %s, got %s\n", "resource_id", r.Id)
	}
	if r.Name != "resource_name" {
		t.Fatalf("expected %s, got %s\n", "resource_name", r.Name)
	}
}

func Test_NewGlobalResource(t *testing.T) {

	ctx := getContext()

	object := map[string]interface{}{
		"Foo": "Bar",
		"Tags": map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		},
	}
	r := NewGlobal(ctx, IamRole, "resource_id", "resource_name", object)
	if r.Region != "aws-global" {
		t.Fatalf("expected aws-global, got %s", r.Region)
	}
	if r.Tags["tag1"] != "value1" {
		t.Fatalf("expected tag that was not present\n")
	}
}

func Test_NewResourceVersion(t *testing.T) {

	ctx := getContext()

	object := map[string]interface{}{
		"Foo": "Bar",
		"Tags": map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		},
	}
	r := NewVersion(ctx, IamRole, "resource_id", "resource_name", "v1", object)
	if r.Version != "v1" {
		t.Fatalf("expected v1, got %s\n", r.Version)
	}
	if r.Tags["tag1"] != "value1" {
		t.Fatalf("expected tag that was not present\n")
	}
}

func Test_ResourceAddRelation(t *testing.T) {

	ctx := getContext()
	object := map[string]interface{}{
		"Foo": "Bar",
		"Tags": map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		},
	}
	r := New(ctx, Ec2Instance, "resource_id", "resource_name", object)
	r.AddRelation(IamRole, "role1", "role1")
	r.AddARNRelation(IamRole, "arn:aws:iam::123456789:role/role2")

	if len(r.Relations) != 2 {
		t.Fatalf("expected 2 relationships, got %d\n", len(r.Relations))
	}
	if r.Relations[0].Id != "role1" {
		t.Fatalf("expected relationship with id of role1, got %s\n", r.Relations[0].Id)
	}
	if r.Relations[1].Id != "role2" {
		t.Fatalf("expected relationship with id of role2, got %s\n", r.Relations[1].Id)
	}
}

func getContext() context2.AWSetsCtx {
	config := aws.Config{
		Region: "us-east-1",
	}
	return context2.AWSetsCtx{
		AWSCfg:    config,
		AccountId: "123456789",
		Context:   context.Background(),
		Logger:    nil,
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
