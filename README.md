# chromos

Go library fetching signed timestamp from Google Chrome's time service.

The library derives its logic from the reference implementation of `network_time_tracker` in [Google Chromium](https://chromium.googlesource.com/chromium/src/+/refs/heads/main/components/network_time/network_time_tracker.cc).

The time service speaks [Omaha CUP-ECDSA protocol](https://github.com/google/omaha/blob/main/doc/ClientUpdateProtocolEcdsa.md).

Example:
```go
import (
    "github.com/radioegor146/chromos"
    "fmt"
)

resp, err := chromos.Query(chromos.GetGoogleConfig())
if err != nil {
    panic(err)
}

fmt.Printf("Google's time: %s\n", resp.Time)
```

Available servers:
- `chromos.GetGoogleConfig()` - `http://clients2.google.com/time/1/current` (from Google Chrome sources)
- `chromos.GetMicrosoftConfig()` - `http://edge.microsoft.com/browsernetworktime/time/1/current` (from MSEdge reverse engineering)
