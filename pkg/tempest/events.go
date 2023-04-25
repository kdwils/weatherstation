package tempest

type (
	// ListenGroup describe the event type group to listen on
	ListenGroup string

	// Event describes the type of event that has occurred
	Event string
)

const (
	ListenGroupStart       ListenGroup = "listen_start"
	ListenGroupStop        ListenGroup = "listen_stop"
	ListenGroupStartEvents ListenGroup = "listen_start_events"
	ListenGroupStopEvents  ListenGroup = "listen_stop_events"
	ListenGroupRapidStart  ListenGroup = "listen_rapid_start"
	ListenGroupRapidStop   ListenGroup = "listen_rapid_stop"

	EventAck                Event = "ack"
	EventConnectionOpened   Event = "connection_opened"
	EventPrecipitation      Event = "evt_precip"
	EventLightingStrike     Event = "evt_strike"
	EventDeviceOnline       Event = "evt_device_online"
	EventDeviceOffline      Event = "evt_device_offline"
	EventStationOnline      Event = "evt_station_online"
	EventStationOffline     Event = "evt_station_offline"
	EventRapidWind          Event = "rapid_wind"
	EventObservationAir     Event = "obs_air"
	EventObservationSky     Event = "obs_sky"
	EventObservationTempest Event = "obs_st"
)
