package collector

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-kit/log"
)

func TestNewClassCollector(t *testing.T) {
	nn := "node1"
	metrics1 := noMetrics
	metrics2 := allMetrics
	metrics3 := "mb,llc,cpu"
	type fields struct {
		classCollectorMetrics *string
		nodeName              *string
	}
	type args struct {
		logger   log.Logger
		interval time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Collector
		wantErr bool
	}{
		{
			name: "TestNewNodeCollector 1",
			fields: fields{
				classCollectorMetrics: &metrics1,
				nodeName:              &nn,
			},
			args: args{
				logger:   log.NewNopLogger(),
				interval: 3 * time.Second,
			},
			want: &classCollector{
				statsCache: make(map[string]*stats),
				interval:   3 * time.Second,
				logger:     log.NewNopLogger(),
				nodeName:   "node1",
				metrics:    make(map[string]struct{}),
			},
			wantErr: false,
		},
		{
			name: "TestNewNodeCollector 2",
			fields: fields{
				classCollectorMetrics: &metrics2,
				nodeName:              &nn,
			},
			args: args{
				logger:   log.NewNopLogger(),
				interval: 3 * time.Second,
			},
			want: &classCollector{
				statsCache: make(map[string]*stats),
				interval:   3 * time.Second,
				logger:     log.NewNopLogger(),
				nodeName:   "node1",
				metrics: map[string]struct{}{
					"mb":  {},
					"llc": {},
				},
			},
			wantErr: false,
		},
		{
			name: "TestNewNodeCollector 3",
			fields: fields{
				classCollectorMetrics: &metrics3,
				nodeName:              &nn,
			},
			args: args{
				logger:   log.NewNopLogger(),
				interval: 3 * time.Second,
			},
			want: &classCollector{
				statsCache: make(map[string]*stats),
				interval:   3 * time.Second,
				logger:     log.NewNopLogger(),
				nodeName:   "node1",
				metrics: map[string]struct{}{
					"mb":  {},
					"llc": {},
					"cpu": {},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classCollectorMetrics = tt.fields.classCollectorMetrics
			nodeName = tt.fields.nodeName
			got, err := NewClassCollector(tt.args.logger, tt.args.interval)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClassCollector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClassCollector() = %v, want %v", got, tt.want)
			}
		})
	}
}
