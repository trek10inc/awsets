module github.com/trek10inc/awsets/cmd/awsets

go 1.15

require (
	github.com/trek10inc/awsets v.event.release.tag_name
	github.com/aws/aws-sdk-go-v2 v0.24.0
	github.com/emicklei/dot v0.11.0
	github.com/jmespath/go-jmespath v0.3.0
	github.com/urfave/cli/v2 v2.2.0
	go.etcd.io/bbolt v1.3.5
)

replace github.com/trek10inc/awsets => ../..
