# Weatherstation
 
Weatherstation is a Go package for consuming real-time event data from a [Tempest weather device](https://apidocs.tempestwx.com/reference/quick-start).

The binary ships with two interfaces:
* A terminal UI
* A simple web dashboard

## How it works

```mermaid
sequenceDiagram
    participant User
    participant Interface
    participant WeatherDevice

    User->>Interface: Launches TUI
    Interface->>WeatherDevice: Subscribes to real-time data
    WeatherDevice-->>Interface: Sends observation data
    Interface->>Interface: Append history metrics
    Interface->>Interface: Render updated graphs
    Interface-->>User: Display updated interface
```

## Installation

To install the binary, run:
From Go
```bash
go install github.com/kdwils/weatherstation@latest
```

Environment Variables:

WebSocket API Reference: https://weatherflow.github.io/Tempest/api/ws.html
The following environment variables are required to connect to the Tempest WebSocket API:
```shell
export WEATHERSTATION_TEMPEST_DEVICE_ID='<your-device-id>'
export WEATHERSTATION_TEMPEST_TOKEN='<your-token>'
export WEATHERSTATION_TEMPEST_SCHEME='wss'
export WEATHERSTATION_TEMPEST_PATH='/swd/data'
export WEATHERSTATION_TEMPEST_HOST='ws.weatherflow.com'
```

To attempt a UDP connection, set the following environment variables:
```shell
export WEATHERSTATION_TEMPEST_DEVICE_ID='<your-device-id>'
export WEATHERSTATION_TEMPEST_TOKEN='<your-token>'
export WEATHERSTATION_TEMPEST_PATH='/swd/data'
export WEATHERSTATION_TEMPEST_HOST='<your-device-lan-ip-address>:54000'
export WEATHERSTATION_TEMPEST_SCHEME='udp'
```

> [!WARNING]
> UDP connectivity is currently untested in this project due to remote development without access to a local device. If it doesn't work for you, open an issue and I'll try to help.

This will install the `weatherstation` binary in your `$GOPATH/bin` directory.

## The Terminal UI

To start the terminal UI:
```shell
weatherstation tui
```
![teminal ui](images/tui.png)

## The Dashboard

The dashboard is a simple web application that uses the Go template engine to render the current weather data.

The dashboard is served on port 8080 by default, but can be configured using the `WEATHERSTATION_SERVER_PORT` environment variable.

![alt text](images/dashboard.png)

To serve the dashboard http server:
From the binary:
```bash
weatherstation serve 
```

Then open http://localhost:8080 (or wherever it's hosted) in your browser to view the dashboard.

## Package Structure

The package is organized into several modules under the `pkg` directory:

### api
`/pkg/api/`
- Contains data models and client interfaces for interacting with the Tempest API
- Handles parsing and conversion of weather observation data
- Provides utility functions for unit conversions (m/s to mph, celsius to fahrenheit, etc.)

### connection
`/pkg/connection/`
- Provides abstract connection interfaces for different protocols
- Implements both WebSocket and UDP connections
- Handles connection lifecycle (connect, read, write, close)

### tempest
`/pkg/tempest/`
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
    deviceID := 123
    conn, err := connection.NewConnection(ctx, "wss", "ws.weatherflow.com", "/swd/data", "your-token")
    if err != nil {
        log.Fatal(err)
    }
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

## Supported Events

The package supports the following event types (defined in `pkg/tempest/events.go`):

- `EventConnectionOpened`: Connection establishment
- `EventObservationTempest`: Weather observations
- `EventPrecipitation`: Precipitation events
- `EventLightingStrike`: Lightning detection
- `EventDeviceOnline`/`EventDeviceOffline`: Device status
- `EventStationOnline`/`EventStationOffline`: Station status
- `EventRapidWind`: Rapid wind measurements

## Acknowledgements
* [go-asciigraph](https://github.com/guptarohit/asciigraph) — for rendering terminal graphs.
* [go-lipgloss](https://github.com/charmbracelet/lipgloss) — for styling the terminal UI.