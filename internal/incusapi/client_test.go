package incusapi

import (
	"testing"
)

func TestClient(t *testing.T) {
	_, err := NewClient()
	if err != nil {
		t.Fatalf("Client() failed: %v", err)
	}
}
