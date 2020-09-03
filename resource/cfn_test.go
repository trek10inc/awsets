package resource

import (
	"errors"
	"reflect"
	"testing"
)

func Test_FromCfn(t *testing.T) {
	tests := map[string]struct {
		res string
		err error
	}{
		"Not::A::Resource": {
			res: "Not/A/Resource",
			err: errors.New("no type defined for Not::A::Resource"),
		},
		"AWS::EC2::Instance": {
			res: Ec2Instance.String(),
			err: nil,
		},
	}
	for k, v := range tests {
		t.Run(k, func(t *testing.T) {
			r, err := FromCfn(k)
			if !reflect.DeepEqual(err, v.err) || r.String() != v.res {
				t.Errorf("expected %s-%v, got %s-%v", v.res, v.err, r, err)
			}
		})
	}
}
