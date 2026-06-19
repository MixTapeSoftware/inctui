package incusapi

import (
	"testing"
)

func TestInstances(t *testing.T) {
	fetcher, err := NewInstanceFetcher()
	if err != nil {
		t.Fatalf("NewInstanceFetcher() failed: %v", err)
	}

	got, err := fetcher.Instances()
	for _, instance := range got {

		t.Logf("%v\n", instance.Name)
	}
	if err != nil {
		t.Fatalf("Instances() failed: %v", err)
	}
}
