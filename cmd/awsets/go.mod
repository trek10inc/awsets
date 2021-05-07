module github.com/trek10inc/awsets/cmd/awsets

go 1.15

require (
	github.com/aws/aws-sdk-go-v2 v1.4.0
	github.com/aws/aws-sdk-go-v2/config v1.1.7
	github.com/cheggaaa/pb/v3 v3.0.8
	github.com/emicklei/dot v0.15.0
	github.com/jmespath/go-jmespath v0.4.0
	github.com/trek10inc/awsets v0.9.0
	github.com/urfave/cli/v2 v2.3.0
	go.etcd.io/bbolt v1.3.5
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)

replace github.com/trek10inc/awsets => ../..
