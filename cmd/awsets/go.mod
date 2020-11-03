module github.com/trek10inc/awsets/cmd/awsets

go 1.15

require (
	github.com/aws/aws-sdk-go-v2 v0.29.0
	github.com/aws/aws-sdk-go-v2/config v0.2.0
	github.com/cheggaaa/pb/v3 v3.0.5
	github.com/emicklei/dot v0.14.0
	github.com/jmespath/go-jmespath v0.4.0
	github.com/trek10inc/awsets v0.6.0
	github.com/urfave/cli/v2 v2.2.0
	go.etcd.io/bbolt v1.3.5
)

replace github.com/trek10inc/awsets => ../..
