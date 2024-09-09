package collector

import (
	"sync"
	"testing"
	"time"

	"github.com/go-kit/log"
)

func TestParseCollectorMetrics(t *testing.T) {
	type fields struct {
		containerCollectorMetrics *string
		classCollectorMetrics     *string
		nodeCollectorMetrics      *string
	}
	metrics1 := noMetrics
	metrics2 := allMetrics
	metrics3 := "mb"
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "TestParseCollectorMetrics 1",
			fields: fields{
				containerCollectorMetrics: nil,
				classCollectorMetrics:     nil,
				nodeCollectorMetrics:      nil,
			},
			want: false,
		},
		{
			name: "TestParseCollectorMetrics 2",
			fields: fields{
				containerCollectorMetrics: &metrics1,
				classCollectorMetrics:     nil,
				nodeCollectorMetrics:      nil,
			},
			want: false,
		},
		{
			name: "TestParseCollectorMetrics 3",
			fields: fields{
				containerCollectorMetrics: &metrics2,
				classCollectorMetrics:     nil,
				nodeCollectorMetrics:      nil,
			},
			want: true,
		},
		{
			name: "TestParseCollectorMetrics 4",
			fields: fields{
				containerCollectorMetrics: &metrics3,
				classCollectorMetrics:     nil,
				nodeCollectorMetrics:      nil,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			containerCollectorMetrics = tt.fields.containerCollectorMetrics
			classCollectorMetrics = tt.fields.classCollectorMetrics
			nodeCollectorMetrics = tt.fields.nodeCollectorMetrics
			if got := ParseCollectorMetrics(); got != tt.want {
				t.Errorf("ParseCollectorMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewCollector(t *testing.T) {
	isDefaultEnabled := true
	isDefaultDisabled := false
	type fields struct {
		collectorState map[string]*bool
	}
	type args struct {
		logger   log.Logger
		interval time.Duration
		filters  []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestNewCollector 1",
			fields: fields{
				collectorState: map[string]*bool{
					containerCollectorSubsystem: &isDefaultEnabled,
					classCollectorSubsystem:     &isDefaultDisabled,
					nodeCollectorSubsystem:      &isDefaultDisabled,
				},
			},
			args: args{
				logger:   log.NewNopLogger(),
				interval: 3 * time.Second,
				filters:  []string{},
			},
			wantErr: false,
		},
		{
			name: "TestNewCollector 2",
			fields: fields{
				collectorState: map[string]*bool{
					// containerCollectorSubsystem: &isDefaultDisabled,
					classCollectorSubsystem: &isDefaultDisabled,
					nodeCollectorSubsystem:  &isDefaultDisabled,
				},
			},
			args: args{
				logger:   log.NewNopLogger(),
				interval: 3 * time.Second,
				filters: []string{
					containerCollectorSubsystem,
				},
			},
			wantErr: true,
		},
		{
			name: "TestNewCollector32",
			fields: fields{
				collectorState: map[string]*bool{
					containerCollectorSubsystem: &isDefaultDisabled,
					classCollectorSubsystem:     &isDefaultDisabled,
					nodeCollectorSubsystem:      &isDefaultDisabled,
				},
			},
			args: args{
				logger:   log.NewNopLogger(),
				interval: 3 * time.Second,
				filters: []string{
					containerCollectorSubsystem,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				collectorState = tt.fields.collectorState
				initiatedCollectorsMtx = sync.Mutex{}
				initiatedCollectors = make(map[string]Collector)
				_, err := NewCollector(tt.args.logger, tt.args.interval, tt.args.filters...)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewCollector() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			},
		)
	}
}

func TestIsNoDataError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "TestIsNoDataError 1",
			args: args{
				err: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNoDataError(tt.args.err); got != tt.want {
				t.Errorf("IsNoDataError() = %v, want %v", got, tt.want)
			}
		})
	}
}
