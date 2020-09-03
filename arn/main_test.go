package arn

import "testing"

func Test_IsArn(t *testing.T) {
	tests := map[string]bool{
		"boop":     false,
		"arn":      false,
		"":         false,
		"arn:boop": true,
	}
	for k, v := range tests {
		t.Run(k, func(t *testing.T) {
			if IsArn(k) != v {
				t.Fail()
			}
		})
	}
}

func stringPointer(input string) *string {
	return &input
}

func Test_IsArnP(t *testing.T) {
	tests := map[*string]bool{
		nil:                       false,
		stringPointer(""):         false,
		stringPointer("arn"):      false,
		stringPointer("arn:boop"): true,
	}
	for k, v := range tests {
		t.Run("", func(t *testing.T) {
			if IsArnP(k) != v {
				t.Fail()
			}
		})
	}
}

//TODO: Parsing tests
