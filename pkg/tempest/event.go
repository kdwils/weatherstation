package tempest

type (
	// ListenEventType describe an event type to listen on
	ListenEventType string

	// EventType describes the type of event that has occurred
	EventType string
)

const (
	ListenStart       ListenEventType = "listen_start"
	ListenStop        ListenEventType = "listen_stop"
	ListenStartEvents ListenEventType = "listen_start_events"
	ListenStopEvents  ListenEventType = "listen_stop_events"
	ListenRapidStart  ListenEventType = "listen_rapid_start"
	ListenRapidStop   ListenEventType = "listen_rapid_stop"

	EventAck                EventType = "ack"
	EventConnectionOpened   EventType = "connection_opened"
	EventPrecipitation      EventType = "evt_precip"
	EventLightingStrike     EventType = "evt_strike"
	EventDeviceOnline       EventType = "evt_device_online"
	EventDeviceOffline      EventType = "evt_device_offline"
	EventStationOnline      EventType = "evt_station_online"
	EventStationOffline     EventType = "evt_station_offline"
	EventRapidWind          EventType = "rapid_wind"
	EventObservationAir     EventType = "obs_air"
	EventObservationSky     EventType = "obs_sky"
	EventObservationTempest EventType = "obs_st"
)
