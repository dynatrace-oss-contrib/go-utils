package config

import (
	"reflect"
	"testing"
)

func TestGetKeptnGoUtilsConfig(t *testing.T) {
	tests := []struct {
		name    string
		want    *KeptnGoUtilsConfig
		wantErr bool
	}{
		{
			name:    "get config",
			want:    &KeptnGoUtilsConfig{ShKeptnSpecVersion: "0.2.3"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetKeptnGoUtilsConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKeptnGoUtilsConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetKeptnGoUtilsConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
