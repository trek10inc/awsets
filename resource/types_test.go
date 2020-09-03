package resource

import "testing"

func Test_ResourceTypeStringer(t *testing.T) {
	tests := map[ResourceType]string{
		Unmapped:       "unmapped",
		Unnecessary:    "unnecessary",
		AcmCertificate: "acm/certificate",
	}
	for k, v := range tests {
		t.Run(k.String(), func(t *testing.T) {
			if v != k.String() {
				t.Errorf("wanted %s, got %s\n", v, k.String())
			}
		})
	}
}
