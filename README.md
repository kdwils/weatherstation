# Weatherstation

A Go package for listening to Tempest weather station events. This package provides a simple interface for consuming real-time weather data from your Tempest weather station.

## Package Structure

The package is organized into several modules under the `pkg` directory:

### api
- Contains data models and client interfaces for interacting with the Tempest API
- Handles parsing and conversion of weather observation data
- Provides utility functions for unit conversions (m/s to mph, celsius to fahrenheit, etc.)

### connection
- Provides abstract connection interfaces for different protocols
- Implements both WebSocket and UDP connections
- Handles connection lifecycle (connect, read, write, close)

### tempest
- Core event listening functionality
- Event type definitions and constants
- Handler registration for different event types

## Usage

Here's an example of how to use the package to listen for weather station events:
A similar one can be found in the [`cmd/listen.go`](https://github.com/kdwils/weatherstation/blob/main/cmd/listen.go) file.
```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "os"
    "os/signal"

    "github.com/kdwils/weatherstation/pkg/api"
    "github.com/kdwils/weatherstation/pkg/connection"
    "github.com/kdwils/weatherstation/pkg/tempest"
)

func main() {
    conn, err := connection.NewConnection(ctx, "wss", "ws.weatherflow.com", "/swd/data", "your-token")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    listener := tempest.NewEventListener(conn, tempest.ListenGroupStart, deviceID)

    // Register handlers for different events
    listener.RegisterHandler(tempest.EventConnectionOpened, func(ctx context.Context, b []byte) {
        log.Printf("connection opened: %s", b)
    })

    listener.RegisterHandler(tempest.EventObservationTempest, func(ctx context.Context, b []byte) {
        var obs api.ObservationTempest
        if err := json.Unmarshal(b, &obs); err != nil {
            log.Fatal(err)
            return
        }
        log.Printf("received observation: %+v", obs)
    })

    // Start listening in a goroutine
    go func() {
        if err := listener.Listen(ctx); err != nil {
            log.Fatal(err)
        }
    }()

    // Wait for interrupt signal
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    <-c
}
```

## Configuration

The package uses [Viper](https://github.com/spf13/viper) for configuration management. You can configure the package using environment variables or a configuration file.

### Environment Variables

```bash
WEATHERSTATION_TEMPEST_HOST=ws.weatherflow.com
WEATHERSTATION_TEMPEST_SCHEME=wss
WEATHERSTATION_TEMPEST_PATH=/swd/data
WEATHERSTATION_TEMPEST_TOKEN=your-token-here
WEATHERSTATION_TEMPEST_DEVICE_ID=your-device-id
```

### Configuration File (yaml)

Create a `config.yaml` file:

```yaml
WEATHERSTATION_TEMPEST_HOST: ws.weatherflow.com
WEATHERSTATION_TEMPEST_SCHEME: wss
WEATHERSTATION_TEMPEST_PATH: /swd/data
WEATHERSTATION_TEMPEST_TOKEN: your-token-here
WEATHERSTATION_TEMPEST_DEVICE_ID: your-device-id
```

### Default Configuration Values

The CLI sets these default values if not otherwise specified:

```go
viper.SetDefault("tempest.host", "")
viper.SetDefault("tempest.scheme", "")
viper.SetDefault("tempest.uri_path", "")
viper.SetDefault("tempest.token", "")
viper.SetDefault("tempest.deviceID", 0)
```

## Supported Events

The package supports various event types defined in `pkg/tempest/events.go`:

- `EventConnectionOpened`: Connection establishment
- `EventObservationTempest`: Weather observations
- `EventPrecipitation`: Precipitation events
- `EventLightingStrike`: Lightning detection
- `EventDeviceOnline`/`EventDeviceOffline`: Device status
- `EventStationOnline`/`EventStationOffline`: Station status
- `EventRapidWind`: Rapid wind measurements