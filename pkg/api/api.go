package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
)

type Client interface {
	GetStationMetadata(ctx context.Context, token string) (StationMetadata, error)
	GetLatestStationObservation(ctx context.Context, stationID, token string) (ObservationReport, error)
	GetLatestDeviceObservation(ctx context.Context, deviceID, token string) (ObservationTempest, error)
}

type Status struct {
	StatusMessage string `json:"status_message"`
	StatusCode    int    `json:"status_code"`
}

type DeviceMeta struct {
	Environment     string  `json:"environment"`
	Name            string  `json:"name"`
	WifiNetworkName string  `json:"wifi_network_name"`
	Agl             float64 `json:"agl"`
}

type DeviceSettings struct {
	ShowPrecipFinal bool `json:"show_precip_final"`
}

type Device struct {
	DeviceType       string         `json:"device_type"`
	FirmwareRevision string         `json:"firmware_revision"`
	HardwareRevision string         `json:"hardware_revision"`
	SerialNumber     string         `json:"serial_number"`
	DeviceMeta       DeviceMeta     `json:"device_meta"`
	DeviceID         int            `json:"device_id"`
	LocationID       int            `json:"location_id"`
	DeviceSettings   DeviceSettings `json:"device_settings,omitempty"`
}

type StationItem struct {
	Item           string `json:"item"`
	DeviceID       int    `json:"device_id"`
	LocationID     int    `json:"location_id"`
	LocationItemID int    `json:"location_item_id"`
	Sort           int    `json:"sort"`
	StationID      int    `json:"station_id"`
	StationItemID  int    `json:"station_item_id"`
}

type StationMeta struct {
	Elevation   float64 `json:"elevation"`
	ShareWithWf bool    `json:"share_with_wf"`
	ShareWithWu bool    `json:"share_with_wu"`
}

type Station struct {
	Name               string        `json:"name"`
	Timezone           string        `json:"timezone"`
	PublicName         string        `json:"public_name"`
	Devices            []Device      `json:"devices"`
	StationItems       []StationItem `json:"station_items"`
	StationMeta        StationMeta   `json:"station_meta"`
	Longitude          float64       `json:"longitude"`
	CreatedEpoch       int64         `json:"created_epoch"`
	LocationID         int           `json:"location_id"`
	StationID          int           `json:"station_id"`
	Latitude           float64       `json:"latitude"`
	LastModifiedEpoch  int64         `json:"last_modified_epoch"`
	TimezoneOffsetMins int           `json:"timezone_offset_minutes"`
	IsLocalMode        bool          `json:"is_local_mode"`
}

type StationMetadata struct {
	Stations []Station `json:"stations"`
	Status   Status    `json:"status"`
}

type Observations struct {
	PressureTrend                    string  `json:"pressure_trend"`
	PrecipAccumLocalYesterdayFinal   float64 `json:"precip_accum_local_yesterday_final"`
	DeltaT                           float64 `json:"delta_t"`
	PrecipAccumLocalYesterday        float64 `json:"precip_accum_local_yesterday"`
	AirDensity                       float64 `json:"air_density"`
	DewPoint                         float64 `json:"dew_point"`
	FeelsLike                        float64 `json:"feels_like"`
	HeatIndex                        float64 `json:"heat_index"`
	LightningStrikeCount             int     `json:"lightning_strike_count"`
	LightningStrikeCountLast1hr      int     `json:"lightning_strike_count_last_1hr"`
	LightningStrikeCountLast3hr      int     `json:"lightning_strike_count_last_3hr"`
	LightningStrikeLastDistance      int     `json:"lightning_strike_last_distance"`
	LightningStrikeLastEpoch         int64   `json:"lightning_strike_last_epoch"`
	Precip                           float64 `json:"precip"`
	PrecipAccumLast1hr               float64 `json:"precip_accum_last_1hr"`
	PrecipAccumLocalDay              float64 `json:"precip_accum_local_day"`
	PrecipAccumLocalDayFinal         float64 `json:"precip_accum_local_day_final"`
	Brightness                       int     `json:"brightness"`
	PrecipAnalysisTypeYesterday      int     `json:"precip_analysis_type_yesterday"`
	BarometricPressure               float64 `json:"barometric_pressure"`
	PrecipMinutesLocalDay            int     `json:"precip_minutes_local_day"`
	PrecipMinutesLocalYesterday      int     `json:"precip_minutes_local_yesterday"`
	PrecipMinutesLocalYesterdayFinal int     `json:"precip_minutes_local_yesterday_final"`
	AirTemperature                   float64 `json:"air_temperature"`
	RelativeHumidity                 int     `json:"relative_humidity"`
	SeaLevelPressure                 float64 `json:"sea_level_pressure"`
	SolarRadiation                   int     `json:"solar_radiation"`
	StationPressure                  float64 `json:"station_pressure"`
	Timestamp                        int64   `json:"timestamp"`
	UV                               float64 `json:"uv"`
	WetBulbGlobeTemperature          float64 `json:"wet_bulb_globe_temperature"`
	WetBulbTemperature               float64 `json:"wet_bulb_temperature"`
	WindAvg                          float64 `json:"wind_avg"`
	WindChill                        float64 `json:"wind_chill"`
	WindDirection                    int     `json:"wind_direction"`
	WindGust                         float64 `json:"wind_gust"`
	WindLull                         float64 `json:"wind_lull"`
}

type StationUnits struct {
	UnitsDirection string `json:"units_direction"`
	UnitsDistance  string `json:"units_distance"`
	UnitsOther     string `json:"units_other"`
	UnitsPrecip    string `json:"units_precip"`
	UnitsPressure  string `json:"units_pressure"`
	UnitsTemp      string `json:"units_temp"`
	UnitsWind      string `json:"units_wind"`
}

type ObservationReport struct {
	StationUnits StationUnits   `json:"station_units"`
	PublicName   string         `json:"public_name"`
	StationName  string         `json:"station_name"`
	Timezone     string         `json:"timezone"`
	Observations []Observations `json:"obs"`
	OutdoorKeys  []string       `json:"outdoor_keys"`
	Status       Status         `json:"status"`
	Elevation    float64        `json:"elevation"`
	Latitude     float64        `json:"latitude"`
	Longitude    float64        `json:"longitude"`
	StationID    int            `json:"station_id"`
	IsPublic     bool           `json:"is_public"`
}

type Observation struct {
	Type    string `json:"type"`
	Device  int    `json:"device_id,omitempty"`
	Station int    `json:"station_id,omitempty"`
}

// ObservationTempest describes the event payload for a 'obs_st' event
type ObservationTempest struct {
	Status  Status                    `json:"status"`
	Type    string                    `json:"type"`
	Source  string                    `json:"source"`
	Summary ObservationTempestSummary `json:"summary"`
	Data    ObservationTempestData    `json:"obs"`
	Device  int                       `json:"device_id"`
}

type ObservationTempestData struct {
	TimeEpoch                       int     `json:"time_epoch"`
	WindLull                        float64 `json:"wind_lull"`
	WindAverage                     float64 `json:"wind_average"`
	WindGust                        float64 `json:"wind_gust"`
	WindDirectionDegrees            float64 `json:"wind_direction"`
	WindSampleInterval              int     `json:"wind_sample_interval"`
	StationPressure                 float64 `json:"station_pressure"`
	AirTemperature                  float64 `json:"air_temperature"`
	RelativeHumidity                int     `json:"relative_humidity"`
	Illuminance                     int     `json:"illuminance"`
	UltraviolentIndex               float64 `json:"uv_index"`
	SolarRadiation                  int     `json:"solar_radiation"`
	RainAccumulated                 float64 `json:"rain_accumulated"`
	PrecipitationType               int     `json:"precipitation_type"`
	LightningStrikeAverageDistance  float64 `json:"lightning_strike_avg_distance"`
	LightningStrikeCount            int     `json:"lightning_strike_count"`
	BatteryVolts                    float64 `json:"battery_volts"`
	ReportInterval                  int     `json:"report_interval"`
	LocalDailyRainAccumulation      float64 `json:"local_daily_rain_accumulation"`
	RainAccumulationFinalCheck      float64 `json:"rain_accumulation_final_check"`
	LocalRainAccumulationFinalCheck float64 `json:"local_rain_accumulation_final_check"`
	PrecipitationAnalysisType       int     `json:"precipitation_analysis_type"`
}

func (o *ObservationTempestData) UnmarshalJSON(b []byte) error {
	data := make([][]any, 0)
	err := json.Unmarshal(b, &data)
	if err != nil {
		return fmt.Errorf("invalid observation event: %v", err)
	}

	if len(data) < 1 {
		return fmt.Errorf("no observation data in payload")
	}

	const (
		totalObservationFields = 22
	)

	obs := data[0]
	if len(obs) != totalObservationFields {
		return fmt.Errorf("observation data is missing: %d total, expected %d", len(obs), totalObservationFields)
	}

	o.TimeEpoch = int(obs[0].(float64))
	o.WindLull = obs[1].(float64)
	o.WindAverage = obs[2].(float64)
	o.WindGust = obs[3].(float64)
	o.WindDirectionDegrees = obs[4].(float64)
	o.WindSampleInterval = int(obs[5].(float64))
	o.StationPressure = obs[6].(float64)
	o.AirTemperature = obs[7].(float64)
	o.RelativeHumidity = int(obs[8].(float64))
	o.Illuminance = int(obs[9].(float64))
	o.UltraviolentIndex = obs[10].(float64)
	o.SolarRadiation = int(obs[11].(float64))
	o.RainAccumulated = obs[12].(float64)
	o.PrecipitationType = int(obs[13].(float64))
	o.LightningStrikeAverageDistance = obs[14].(float64)
	o.LightningStrikeCount = int(obs[15].(float64))
	o.BatteryVolts = obs[16].(float64)
	o.ReportInterval = int(obs[17].(float64))
	o.LocalDailyRainAccumulation = obs[18].(float64)
	o.RainAccumulationFinalCheck = obs[19].(float64)
	o.LocalRainAccumulationFinalCheck = obs[20].(float64)
	o.PrecipitationAnalysisType = int(obs[21].(float64))
	return nil
}

type ObservationTempestSummary struct {
	PressureTrend                  string  `json:"pressure_trend"`
	StrikeCountOneHour             int     `json:"strike_count_1h"`
	StrikeCountThreeHour           int     `json:"strike_count_3h"`
	PrecipTotalOneHour             float64 `json:"precip_total_1h"`
	StrikeLastDistance             int     `json:"strike_last_dist"`
	StrikeLastEpoch                int     `json:"strike_last_epoch"`
	PrecipAccumLocalYesterday      float64 `json:"precip_accum_local_yesterday"`
	PrecipAccumLocalYesterdayFinal float64 `json:"precip_accum_local_yesterday_final"`
	FeelsLike                      float64 `json:"feels_like"`
	HeatIndex                      float64 `json:"heat_index"`
	WindChill                      float64 `json:"wind_chill"`
	DewPoint                       float64 `json:"dew_point"`
	WetBulbTemperature             float64 `json:"web_bulb_temperature"`
	AirDensity                     float64 `json:"air_density"`
	DeltaT                         float64 `json:"delta_t"`
	PrecipMinutesLocalDay          int     `json:"precip_minutes_local_day"`
	PrecipMinutesLocalYesterday    int     `json:"precip_minutes_local_yesterday"`
}

func (o ObservationTempest) IsRaining() bool {
	return o.Data.PrecipitationAnalysisType == 1 || o.Data.PrecipitationAnalysisType == 2
}

func (o ObservationTempest) WindDirection() string {
	var (
		compassPoints = []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	)

	degrees := math.Mod((o.Data.WindDirectionDegrees + 360), 360)
	degreeStep := 360.0 / float64(len(compassPoints))
	index := int((degrees/degreeStep)+0.5) % len(compassPoints)
	return compassPoints[index]
}

func (o ObservationTempest) WindSpeedGustMPH() float64 {
	return metersPerSecondToMilesPerHour(o.Data.WindGust)
}

func (o ObservationTempest) WindSpeedAverageMPH() float64 {
	return metersPerSecondToMilesPerHour(o.Data.WindAverage)
}

func (o ObservationTempest) RainfallInInches() float64 {
	return millimetersToInches(o.Data.RainAccumulated)
}

func (o ObservationTempest) RainfallYesterdayInInches() float64 {
	return millimetersToInches(float64(o.Summary.PrecipMinutesLocalYesterday))
}

func (o ObservationTempest) TemperatureInFarneheit() float64 {
	return celsiusToFahrenheit(o.Data.AirTemperature)
}

func (o ObservationTempest) FeelsLikeFarenheit() float64 {
	return celsiusToFahrenheit(o.Summary.FeelsLike)
}

func (o ObservationTempest) DewPointFarenheit() float64 {
	return celsiusToFahrenheit(o.Summary.DewPoint)
}

func (o ObservationTempest) PrecipitationType() string {
	switch o.Data.PrecipitationType {
	case 1:
		return "Raining"
	case 2:
		return "Hailing"
	default:
		return "Dry"
	}
}

func (o ObservationTempest) AverageLightningStrikeDistanceInMiles() float64 {
	return kilometersToMiles(o.Data.LightningStrikeAverageDistance)
}

func metersPerSecondToMilesPerHour(mps float64) float64 {
	const conversion = 2.23694
	return mps * conversion
}

func millimetersToInches(mm float64) float64 {
	const conversion = 0.03937
	return mm * conversion
}

func celsiusToFahrenheit(celsius float64) float64 {
	return celsius*9/5 + 32
}

func kilometersToMiles(km float64) float64 {
	const conversion = 0.621371
	return km * conversion
}
