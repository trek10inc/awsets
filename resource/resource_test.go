package resource

import (
	ctx2 "context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/context"
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

func getContext() context.AWSetsCtx {
	config := aws.Config{
		Region: "us-east-1",
	}
	return context.AWSetsCtx{
		AWSCfg:    config,
		AccountId: "123456789",
		Context:   ctx2.Background(),
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

func Test_JSON(t *testing.T) {

	rg := NewGroup()

	cfgUsEast1 := getContext()
	cfgUsEast2 := cfgUsEast1.Copy("us-east-2")
	object := map[string]interface{}{
		"Foo": "Bar",
		"Tags": map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		},
	}
	r1 := New(*cfgUsEast2, Ec2Instance, "resource 1", "resource_1", object)
	rg.AddResource(r1)
	r2 := New(*cfgUsEast2, Ec2Image, "resource 2", "resource_2", object)
	rg.AddResource(r2)
	r3 := NewVersion(cfgUsEast1, Ec2Instance, "resource 3", "resource_3", "2", object)
	rg.AddResource(r3)
	r4 := NewVersion(cfgUsEast1, Ec2Instance, "resource 4", "resource_4", "2", object)
	rg.AddResource(r4)

	jsonStr, err := rg.JSON()
	if err != nil {
		t.Fail()
	}
	var resources []Resource
	err = json.Unmarshal([]byte(jsonStr), &resources)
	if err != nil {
		t.Fail()
	}
	if len(resources) != 4 {
		t.Fatalf("expected 4 resources, got %d\n", len(resources))
	}
	// Sort resources in JSON by Account, Type, Region, Id, then Version to allow for consisting diff-ing
	// First sorts by type, ec2/image comes before ec2/instance
	// Then sorts by region, so the us-east-1 resources come before remaining us-east-2 resources
	// Last sorts by Id, so of the us-east-1 resources, 3 < 4
	// So we expect 2, 3, 4, 1
	if resources[0].Name != "resource_2" {
		t.Fail()
	}
	if resources[1].Name != "resource_3" {
		t.Fail()
	}
	if resources[2].Name != "resource_4" {
		t.Fail()
	}
	if resources[3].Name != "resource_1" {
		t.Fail()
	}
	// 2 3 4 1
}
