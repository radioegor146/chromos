package chromos

import (
	"fmt"
	"testing"
)

func TestGoogleServer(t *testing.T) {
	for ver := minGoogleKeyVersion; ver <= maxGoogleKeyVersion; ver++ {
		if getGoogleKey(ver) == nil {
			continue
		}
		time, err := Query(GetGoogleConfigVersion(ver))
		if err != nil {
			t.Errorf("Query(GetGoogleConfig()) returned error: %v", err)
		}
		if testing.Verbose() {
			fmt.Println("GoogleKey", ver, time)
		}
	}
}

func TestMicrosoftServer(t *testing.T) {
	time, err := Query(GetMicrosoftConfig())
	if err != nil {
		t.Errorf("Query(GetMicrosoftConfig()) returned error: %v", err)
		return
	}
	if testing.Verbose() {
		fmt.Println("MircosoftKey", time)
	}
}

func TestReadmeExample(t *testing.T) {
	resp, err := Query(GetGoogleConfig())
	if err != nil {
		panic(err)
	}
	if testing.Verbose() {
		fmt.Printf("Google's time: %s\n", resp.Time)
	}
}
