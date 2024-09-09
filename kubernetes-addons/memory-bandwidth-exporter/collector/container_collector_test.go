package collector

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/opea-project/GenAIInfra/kubernetes-addons/memory-bandwidth-exporter/info"
)

func TestNewContainerCollector(t *testing.T) {
	nn := "node2"
	metrics1 := noMetrics
	metrics2 := allMetrics
	metrics3 := "llc,cpu"
	namespaceWhiteList1 := "system"
	namespaceWhiteList2 := "system,kube-system"
	type fields struct {
		containerCollectorMetrics *string
		nodeName                  *string
		namespaceWhiteList        *string
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
				containerCollectorMetrics: &metrics1,
				nodeName:                  &nn,
				namespaceWhiteList:        &namespaceWhiteList1,
			},
			args: args{
				logger:   log.NewNopLogger(),
				interval: 3 * time.Second,
			},
			want: &containerCollector{
				statsCache:         make(map[string]stats),
				containerInfos:     make(map[string]info.ContainerInfo),
				interval:           3 * time.Second,
				logger:             log.NewNopLogger(),
				namespaceWhiteList: []string{"system"},
				monTimes:           0,
				metrics:            make(map[string]struct{}),
			},
			wantErr: false,
		},
		{
			name: "TestNewNodeCollector 2",
			fields: fields{
				containerCollectorMetrics: &metrics2,
				nodeName:                  &nn,
				namespaceWhiteList:        &namespaceWhiteList2,
			},
			args: args{
				logger:   log.NewNopLogger(),
				interval: 3 * time.Second,
			},
			want: &containerCollector{
				statsCache:         make(map[string]stats),
				containerInfos:     make(map[string]info.ContainerInfo),
				interval:           3 * time.Second,
				logger:             log.NewNopLogger(),
				namespaceWhiteList: []string{"system", "kube-system"},
				monTimes:           0,
				metrics: map[string]struct{}{
					"mb":     {},
					"llc":    {},
					"cpu":    {},
					"memory": {},
				},
			},
			wantErr: false,
		},
		{
			name: "TestNewNodeCollector 3",
			fields: fields{
				containerCollectorMetrics: &metrics3,
				nodeName:                  &nn,
				namespaceWhiteList:        &namespaceWhiteList2,
			},
			args: args{
				logger:   log.NewNopLogger(),
				interval: 3 * time.Second,
			},
			want: &containerCollector{
				statsCache:         make(map[string]stats),
				containerInfos:     make(map[string]info.ContainerInfo),
				interval:           3 * time.Second,
				logger:             log.NewNopLogger(),
				namespaceWhiteList: []string{"system", "kube-system"},
				monTimes:           0,
				metrics: map[string]struct{}{
					"llc": {},
					"cpu": {},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			containerCollectorMetrics = tt.fields.containerCollectorMetrics
			nodeName = tt.fields.nodeName
			namespaceWhiteList = tt.fields.namespaceWhiteList
			got, err := NewContainerCollector(tt.args.logger, tt.args.interval)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewContainerCollector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewContainerCollector() = %v, want %v", got, tt.want)
			}
		})
	}
}
