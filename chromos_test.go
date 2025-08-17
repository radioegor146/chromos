package chromos

import (
	"testing"
)

func TestGoogleServer(t *testing.T) {
	time, err := FetchTime(GetGoogleConfig())
	if err != nil {
		t.Errorf("FetchTime(GetGoogleConfig()) returned error: %v", err)
		return
	}

	if time == 0 {
		t.Errorf("FetchTime(GetGoogleConfig()) returned time == 0")
		return
	}
}