package incusapi

import (
	"testing"
)

func TestInstances(t *testing.T) {
	client, _ := NewClient()

	//want := "foo"
	got, err := Instances(client)
	for _, instance := range got {

		t.Logf("%v\n", instance.Name)
	}
	if err != nil {
		t.Fatalf("Instances() failed: %v", err)
	}

	//	if want != got {
	//		t.Error("want %v got %v", want, got)
	//	}
}

func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}
