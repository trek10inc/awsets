package context

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type AWSetsCtx struct {
	AWSCfg     aws.Config
	AccountId  string
	WorkerId   int
	Context    context.Context
	Lister     string
	StatusChan chan<- StatusUpdate
	TotalJobs  int
}

func (c *AWSetsCtx) Region() string {
	return c.AWSCfg.Region
}

func (c *AWSetsCtx) Copy(region string) *AWSetsCtx {

	cop := &AWSetsCtx{
		AWSCfg:     c.AWSCfg.Copy(),
		AccountId:  c.AccountId,
		Context:    c.Context,
		StatusChan: c.StatusChan,
		Lister:     c.Lister,
		WorkerId:   c.WorkerId,
		TotalJobs:  c.TotalJobs,
	}
	cop.AWSCfg.Region = region
	return cop
}

func (c *AWSetsCtx) SendStatus(statusType StatusType, msg string) {
	if c.StatusChan == nil {
		return
	}
	su := StatusUpdate{
		Type:      statusType,
		Lister:    c.Lister,
		Region:    c.Region(),
		Message:   msg,
		WorkerId:  c.WorkerId,
		TotalJobs: c.TotalJobs,
	}
	c.StatusChan <- su
}

func (c *AWSetsCtx) Close() {
	if c.StatusChan != nil {
		close(c.StatusChan)
	}
}
