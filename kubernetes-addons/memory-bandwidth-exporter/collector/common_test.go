package collector

import (
	"reflect"
	"testing"
)

func Test_isNeedCollectMbLLc(t *testing.T) {
	type args struct {
		metrics map[string]struct{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "need collect mb and llc",
			args: args{
				metrics: map[string]struct{}{
					"mb":  {},
					"llc": {},
				},
			},
			want: true,
		},
		{
			name: "do not need collect mb and llc",
			args: args{
				metrics: map[string]struct{}{
					"cpu": {},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNeedCollectMbLLc(tt.args.metrics); got != tt.want {
				t.Errorf("isNeedCollectMbLLc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isNeedCollectCpu(t *testing.T) {
	type args struct {
		metrics map[string]struct{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "need collect cpu",
			args: args{
				metrics: map[string]struct{}{
					"mb":  {},
					"llc": {},
					"cpu": {},
				},
			},
			want: true,
		},
		{
			name: "do not need collect cpu",
			args: args{
				metrics: map[string]struct{}{
					"mb":  {},
					"llc": {},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNeedCollectCpu(tt.args.metrics); got != tt.want {
				t.Errorf("isNeedCollectCpu() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isNeedCollectMemory(t *testing.T) {
	type args struct {
		metrics map[string]struct{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "need collect memory",
			args: args{
				metrics: map[string]struct{}{
					"mb":     {},
					"llc":    {},
					"cpu":    {},
					"memory": {},
				},
			},
			want: true,
		},
		{
			name: "do not need collect memory",
			args: args{
				metrics: map[string]struct{}{
					"mb":  {},
					"llc": {},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNeedCollectMemory(tt.args.metrics); got != tt.want {
				t.Errorf("isNeedCollectMemory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getMetricsKeys(t *testing.T) {
	type args struct {
		m map[string]struct{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "get mertics keys 1",
			args: args{
				m: map[string]struct{}{
					"mb":  {},
					"llc": {},
				},
			},
			want: "mb,llc",
		},
		{
			name: "get mertics keys 2",
			args: args{
				m: map[string]struct{}{},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMetricsKeys(tt.args.m); got != tt.want {
				t.Errorf("getMetricsKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bytesToMiB(t *testing.T) {
	type args struct {
		bytes uint64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "bytes to MiB",
			args: args{
				bytes: 1048576,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bytesToMiB(tt.args.bytes); got != tt.want {
				t.Errorf("bytesToMiB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bytesToMB(t *testing.T) {
	type args struct {
		bytes uint64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "bytes to MB",
			args: args{
				bytes: 1000000,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bytesToMB(tt.args.bytes); got != tt.want {
				t.Errorf("bytesToMB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_processStats(t *testing.T) {
	type args struct {
		oldStats RawStats
		newStats RawStats
	}
	tests := []struct {
		name    string
		args    args
		want    ProcessedStats
		wantErr bool
	}{
		{
			name: "process stats",
			args: args{
				oldStats: RawStats{
					SocketNum: 2,
					MemoryBandwidth: map[string]RawMemoryBandwidthStats{
						"0": {
							TotalBytes:          10000000,
							LocalBytes:          5000000,
							TotalBytesTimeStamp: "2021-01-01  15:04:05",
							LocalBytesTimeStamp: "2021-01-01  15:04:05",
						},
						"1": {
							TotalBytes:          10000000,
							LocalBytes:          5000000,
							TotalBytesTimeStamp: "2021-01-01  15:04:05",
							LocalBytesTimeStamp: "2021-01-01  15:04:05",
						},
					},
					Cache: map[string]RawCacheStats{
						"0": {
							LLCOccupancy: 1048576,
						},
						"1": {
							LLCOccupancy: 524288,
						},
					},
					CPUUtilization: &RawCPUStats{
						CPU:       1000000,
						TimeStamp: "2021-01-01  15:04:05",
					},
					Memory: 100,
				},
				newStats: RawStats{
					SocketNum: 2,
					MemoryBandwidth: map[string]RawMemoryBandwidthStats{
						"0": {
							TotalBytes:          20000000,
							LocalBytes:          10000000,
							TotalBytesTimeStamp: "2021-01-01  15:04:15",
							LocalBytesTimeStamp: "2021-01-01  15:04:15",
						},
						"1": {
							TotalBytes:          30000000,
							LocalBytes:          10000000,
							TotalBytesTimeStamp: "2021-01-01  15:04:15",
							LocalBytesTimeStamp: "2021-01-01  15:04:15",
						},
					},
					Cache: map[string]RawCacheStats{
						"0": {
							LLCOccupancy: 1048576,
						},
						"1": {
							LLCOccupancy: 524288,
						},
					},
					CPUUtilization: &RawCPUStats{
						CPU:       2000000,
						TimeStamp: "2021-01-01  15:04:15",
					},
					Memory: 1024 * 1024,
				},
			},
			want: ProcessedStats{
				socketNum: 2,
				SumMemoryBandwidth: ProcessedMemoryBandwidthStats{
					TotalMBps: 3,
					LocalMBps: 1,
				},
				SumCache: ProcessedCacheStats{
					LLCOccupancy: 1.5,
				},
				MemoryBandwidth: map[string]ProcessedMemoryBandwidthStats{
					"0": {
						TotalMBps: 1,
						LocalMBps: 0.5,
					},
					"1": {
						TotalMBps: 2,
						LocalMBps: 0.5,
					},
				},
				Cache: map[string]ProcessedCacheStats{
					"0": {
						LLCOccupancy: 1,
					},
					"1": {
						LLCOccupancy: 0.5,
					},
				},
				CPUUtilization: 0.1,
				Memory:         1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processStats(tt.args.oldStats, tt.args.newStats)
			if (err != nil) != tt.wantErr {
				t.Errorf("processStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processStats() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stringInSlice(t *testing.T) {
	type args struct {
		str  string
		list []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "string in slice",
			args: args{
				str:  "a",
				list: []string{"a", "b", "c"},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringInSlice(tt.args.str, tt.args.list); got != tt.want {
				t.Errorf("stringInSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
