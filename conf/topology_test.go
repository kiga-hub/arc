package conf

import (
	"encoding/json"
	"testing"
)

func TestGenerateFakeTopology(t *testing.T) {
	tests := []struct {
		name    string
		want    *TopologyConfig
		wantErr bool
	}{
		{
			name:    "try",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateFakeTopology()
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateFakeTopology() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			bs, err := json.Marshal(got)
			if err != nil {
				t.Error(err)
			}
			t.Log(string(bs))
		})
	}
}
