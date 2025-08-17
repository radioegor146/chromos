# chromos

Go library for fetching signed current timestamp from Google Chrome's network_time_tracker

Based on [this code from Google](https://chromium.googlesource.com/chromium/src/+/refs/heads/main/components/network_time/network_time_tracker.cc)

Example:
```go
import (
    "github.com/radioegor146/chromos"
    "fmt"
)

time, err := chromos.FetchTime(chromos.GetGoogleConfig())
if err != nil {
    panic(err)
}

fmt.Printf("current time in milliseconds: %d", time)
```