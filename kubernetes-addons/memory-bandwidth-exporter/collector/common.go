package collector

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
)

const (
	monDataDirName        = "mon_data"
	llcOccupancyFileName  = "llc_occupancy"
	mbmLocalBytesFileName = "mbm_local_bytes"
	mbmTotalBytesFileName = "mbm_total_bytes"
	unavailable           = "Unavailable"
	rootResctrlPath       = "/sys/fs/resctrl"
	cgroupControllerPath  = "/sys/fs/cgroup/cgroup.controllers"
	fmtTime               = "2006-01-02 15:04:05"
	allMetrics            = "all"
	noMetrics             = "none"
)

var (
	nodeName = kingpin.Flag(
		"collector.node.name",
		"Give node name.",
	).Default("").String()
	namespaceWhiteList = kingpin.Flag(
		"collector.container.namespaceWhiteList",
		`Filter out containers whose namespaces belong to the namespace whitelist, 
		namespaces separated by commas, like \"xx,yy,zz\".`,
	).Default("").String()
	monTimes = kingpin.Flag(
		"collector.container.monTimes",
		"Scan the pids of containers created before the exporter starts to prevent the loss of pids.",
	).Default("10").Int()
	containerCollectorMetrics = kingpin.Flag(
		"collector.container.metrics",
		"Enable container collector metrics",
	).Default("all").String()
	classCollectorMetrics = kingpin.Flag(
		"collector.class.metrics",
		"Enable class collector metrics",
	).Default("none").String()
	nodeCollectorMetrics = kingpin.Flag(
		"collector.node.metrics",
		"Enable node collector metrics",
	).Default("none").String()
	allClassMetrics     = []string{"mb", "llc"}
	allNodeMetrics      = []string{"mb", "llc"}
	allContainerMetrics = []string{"mb", "llc", "cpu", "memory"}
)

func isNeedCollectMbLLc(metrics map[string]struct{}) bool {
	_, ok1 := metrics["mb"]
	_, ok2 := metrics["llc"]
	return ok1 || ok2
}

func isNeedCollectCpu(metrics map[string]struct{}) bool {
	_, ok := metrics["cpu"]
	return ok
}

func isNeedCollectMemory(metrics map[string]struct{}) bool {
	_, ok := metrics["memory"]
	return ok
}

func getMetricsKeys(m map[string]struct{}) string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return strings.Join(keys, ",")
}

func jitter(duration time.Duration) time.Duration {
	const maxFactor = 0.1
	wait := duration + time.Duration(rand.Float64()*maxFactor*float64(duration))
	return wait
}

// path: mon_groups path
func getIntelRDTStatsFrom(path string) (int, map[string]RawMemoryBandwidthStats, map[string]RawCacheStats, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return 0, nil, nil, fmt.Errorf("mon_groups path %q does not exist", path)
	}
	statsDirectories, err := filepath.Glob(filepath.Join(path, monDataDirName, "*"))
	if err != nil {
		return 0, nil, nil, err
	}

	if len(statsDirectories) == 0 {
		return 0, nil, nil, fmt.Errorf("there is no mon_data stats directories: %q", path)
	}

	cmtStats := make(map[string]RawCacheStats, 0)
	mbmStats := make(map[string]RawMemoryBandwidthStats, 0)

	socketNum := len(statsDirectories)
	for _, dir := range statsDirectories {
		dirParts := strings.Split(dir, "_")
		nid := dirParts[len(dirParts)-1]

		llcOccupancy, _, err := readStatFrom(filepath.Join(dir, llcOccupancyFileName))
		if err != nil {
			return socketNum, nil, nil, err
		}
		cmtStats[nid] = RawCacheStats{
			LLCOccupancy: llcOccupancy,
		}

		totalBytes, tBtime, err := readStatFrom(filepath.Join(dir, mbmTotalBytesFileName))
		if err != nil {
			return socketNum, nil, nil, err
		}
		localBytes, lBtime, err := readStatFrom(filepath.Join(dir, mbmLocalBytesFileName))
		if err != nil {
			return socketNum, nil, nil, err
		}
		mbmStats[nid] = RawMemoryBandwidthStats{
			TotalBytes:          totalBytes,
			TotalBytesTimeStamp: tBtime,
			LocalBytes:          localBytes,
			LocalBytesTimeStamp: lBtime,
		}
	}
	return socketNum, mbmStats, cmtStats, nil
}

// path: cgroupPath
func getCPUUtilizationFrom(path string) (RawCPUStats, error) {
	cgroupVersion := getCgroupVersion()
	var err error
	stat := RawCPUStats{
		TimeStamp: time.Now().Format(fmtTime),
	}
	if cgroupVersion == "v1" {
		stat.CPU, err = getCgroupV1CpuTime(path)
		if err != nil {
			return stat, err
		}
	} else {
		stat.CPU, err = getCgroupV2CpuTime(path)
		if err != nil {
			return stat, err
		}
	}

	return stat, nil
}

// path: cgroupPath
func getMemorySizeFrom(path string) (int64, error) {
	cgroupVersion := getCgroupVersion()
	var err error
	var filePath string
	if cgroupVersion == "v1" {
		filePath = filepath.Join(path, "memory.usage_in_bytes")
	}
	if cgroupVersion == "v2" {
		filePath = filepath.Join(path, "memory.current")
	}
	_, err = os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	content, err := os.ReadFile(filePath)

	if err != nil {
		return 0, err
	}
	memory, err := strconv.ParseInt(strings.TrimSpace(string(content)), 10, 64)
	return memory, err
}

func getCgroupVersion() string {
	_, err := os.Stat(cgroupControllerPath)
	if err == nil {
		return "v2"
	} else {
		return "v1"
	}
}

func getCgroupV1CpuTime(cgroupPath string) (int64, error) {
	filePath := filepath.Join(cgroupPath, "cpuacct.usage")
	_, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}
	cpuUsage, err := strconv.ParseInt(strings.TrimSpace(string(content)), 10, 64)
	return cpuUsage, err
}

// The CPU time obtained is in microseconds.
func getCgroupV2CpuTime(cgroupPath string) (int64, error) {
	filePath := filepath.Join(cgroupPath, "cpu.stat")
	_, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			err = cerr
		}
	}()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		columns := strings.Split(line, " ")
		if columns[0] == "usage_usec" {
			cpuUsage, err := strconv.ParseInt(strings.TrimSpace(columns[1]), 10, 64)
			return cpuUsage, err
		}
	}
	return 0, err
}

// bytesToMiB converts bytes to MiB
func bytesToMiB(bytes uint64) float64 {
	return float64(bytes) / (1024 * 1024)
}

// bytesToMB converts bytes to MB
func bytesToMB(bytes uint64) float64 {
	return float64(bytes) / (1000 * 1000)
}

func readStatFrom(path string) (uint64, string, error) {
	context, err := os.ReadFile(path)
	now := time.Now().Format(fmtTime)
	if err != nil {
		return 0, now, err
	}

	contextString := string(bytes.TrimSpace(context))

	if contextString == unavailable {
		err := fmt.Errorf("\"Unavailable\" value from file %q", path)
		return 0, now, err
	}

	stat, err := strconv.ParseUint(contextString, 10, 64)
	if err != nil {
		return stat, now, fmt.Errorf("unable to parse %q as a uint from file %q", string(context), path)
	}

	return stat, now, nil
}

func processStats(oldStats RawStats, newStats RawStats) (ProcessedStats, error) {
	pstats := ProcessedStats{
		socketNum: newStats.SocketNum,
	}
	var sumCmtStats float64
	var sumMbmTotal float64
	var sumMbmLocal float64

	if newStats.Cache != nil {
		pstats.Cache = make(map[string]ProcessedCacheStats, 0)
		for nid, llc := range newStats.Cache {
			cmt := bytesToMiB(llc.LLCOccupancy)
			sumCmtStats += cmt
			pstats.Cache[nid] = ProcessedCacheStats{
				LLCOccupancy: cmt,
			}
		}
		pstats.SumCache = ProcessedCacheStats{
			LLCOccupancy: sumCmtStats,
		}
	}
	if newStats.MemoryBandwidth != nil && oldStats.MemoryBandwidth != nil {
		pstats.MemoryBandwidth = make(map[string]ProcessedMemoryBandwidthStats, 0)
		for nid, newStat := range newStats.MemoryBandwidth {
			oldStat, ok := oldStats.MemoryBandwidth[nid]
			if !ok {
				return pstats, fmt.Errorf("missing socket %q in oldStats", nid)
			}
			otTime, err := time.Parse(fmtTime, oldStat.TotalBytesTimeStamp)
			if err != nil {
				return pstats, err
			}
			ntTime, err := time.Parse(fmtTime, newStat.TotalBytesTimeStamp)
			if err != nil {
				return pstats, err
			}
			olTime, err := time.Parse(fmtTime, oldStat.LocalBytesTimeStamp)
			if err != nil {
				return pstats, err
			}
			nlTime, err := time.Parse(fmtTime, newStat.LocalBytesTimeStamp)
			if err != nil {
				return pstats, err
			}
			tmbm := bytesToMB(newStat.TotalBytes-oldStat.TotalBytes) / ntTime.Sub(otTime).Seconds()
			lmbm := bytesToMB(newStat.LocalBytes-oldStat.LocalBytes) / nlTime.Sub(olTime).Seconds()
			sumMbmTotal += tmbm
			sumMbmLocal += lmbm
			pstats.MemoryBandwidth[nid] = ProcessedMemoryBandwidthStats{
				TotalMBps: tmbm,
				LocalMBps: lmbm,
			}
		}
		pstats.SumMemoryBandwidth = ProcessedMemoryBandwidthStats{
			TotalMBps: sumMbmTotal,
			LocalMBps: sumMbmLocal,
		}
	}
	if oldStats.CPUUtilization != nil && newStats.CPUUtilization != nil {
		ocTime, err := time.Parse(fmtTime, oldStats.CPUUtilization.TimeStamp)
		if err != nil {
			return pstats, err
		}
		ncTime, err := time.Parse(fmtTime, newStats.CPUUtilization.TimeStamp)
		if err != nil {
			return pstats, err
		}
		pstats.CPUUtilization = float64(newStats.CPUUtilization.CPU-oldStats.CPUUtilization.CPU) /
			float64(ncTime.Sub(ocTime).Microseconds())
	}
	if newStats.Memory != 0 {
		pstats.Memory = float64(newStats.Memory) / 1024 / 1024
	}
	return pstats, nil
}

func makeMonitorGroup(monPath string) error {
	info, err := os.Stat(monPath)
	if os.IsNotExist(err) {
		err := os.Mkdir(monPath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check directory %v: %v", monPath, err)
	} else if !info.IsDir() {
		return fmt.Errorf("%s already exists but is not a directory", monPath)
	} else {
		fmt.Printf("Directory %s already exists\n", monPath)
	}

	return nil
}

func writePidsToTasks(monPath string, cgroupPath string) error {
	containerPids, err := readCPUTasks(cgroupPath + "/cgroup.threads")
	if err != nil {
		return fmt.Errorf("failed to read %v/cgroup.threads: %v", cgroupPath, err)
	}
	err = writeTaskIDsToFile(containerPids, monPath+"/tasks")
	if err != nil {
		return fmt.Errorf("failed to write to %v/tasks: %v", monPath, err)
	}
	return nil
}

func readCPUTasks(path string) ([]int32, error) {
	if path == "" {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	tasksStr := strings.Trim(string(data), "\n")
	values := make([]int32, 0)
	lines := strings.Split(tasksStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) <= 0 {
			continue
		}
		v, err := strconv.ParseInt(line, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("cannot parse cgroup value of line %s, err: %v", line, err)
		}
		values = append(values, int32(v))
	}
	return values, nil
}

func writeTaskIDsToFile(pids []int32, filename string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			err = cerr
		}
	}()
	for _, id := range pids {
		_, err := file.WriteString(strconv.FormatInt(int64(id), 10) + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
