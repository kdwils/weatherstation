package api

import "testing"

func TestObservationTempest_Raining(t *testing.T) {
	type fields struct {
		Type    string
		Source  string
		Status  Status
		Summary ObservationTempestSummary
		Data    ObservationTempestData
		Device  int
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "not raining",
			fields: fields{
				Data: ObservationTempestData{
					PrecipitationAnalysisType: 0,
				},
			},
			want: false,
		},
		{
			name: "raining type 1",
			fields: fields{
				Data: ObservationTempestData{
					PrecipitationAnalysisType: 1,
				},
			},
			want: true,
		},
		{
			name: "raining type 2",
			fields: fields{
				Data: ObservationTempestData{
					PrecipitationAnalysisType: 1,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := ObservationTempest{
				Type:    tt.fields.Type,
				Device:  tt.fields.Device,
				Source:  tt.fields.Source,
				Status:  tt.fields.Status,
				Summary: tt.fields.Summary,
				Data:    tt.fields.Data,
			}
			if got := o.IsRaining(); got != tt.want {
				t.Errorf("ObservationTempest.Raining() = %v, want %v", got, tt.want)
			}
		})
	}
}
