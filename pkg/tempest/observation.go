package tempest

import (
	"encoding/json"
	"fmt"
)

type Status struct {
	Code    int    `json:"status_code"`
	Message string `json:"status_message"`
}

type Observation struct {
	Type    string `json:"type"`
	Device  int    `json:"device_id,omitempty"`
	Station int    `json:"station_id,omitempty"`
}

// ObservationTempest describes the event payload for a 'obs_st' event
type ObservationTempest struct {
	Type    string                    `json:"type"`
	Device  int                       `json:"device_id"`
	Source  string                    `json:"source"`
	Status  Status                    `json:"status"`
	Summary ObservationTempestSummary `json:"summary"`
	Data    ObservationTempestData    `json:"obs"`
}

type ObservationTempestData struct {
	TimeEpoch                       int     `json:"time_epoch"`
	WindLull                        float64 `json:"wind_lull"`
	WindAverage                     float64 `json:"wind_average"`
	WindGust                        float64 `json:"wind_gust"`
	WindDirection                   int     `json:"wind_direction"`
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
	o.WindDirection = int(obs[4].(float64))
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
