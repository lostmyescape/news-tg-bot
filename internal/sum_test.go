package internal

import "testing"

func TestSum(t *testing.T) {
	result := Sum(2, 3)
	if result != 5 {
		t.Errorf("expected 5, got %d", result)
	}
}
