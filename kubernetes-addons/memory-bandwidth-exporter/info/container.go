package info

var (
	chanSize = 100
	// containerid:ContainerInfo
	ContainerInfoChan = make(chan map[string]ContainerInfo, chanSize)
)

type ContainerInfo struct {
	Operation     int // 0:delete 1:add 2:add and watch
	ContainerName string
	ContainerId   string
	PodName       string
	NameSpace     string
	CgroupPath    string
	MonGroupPath  string
}
