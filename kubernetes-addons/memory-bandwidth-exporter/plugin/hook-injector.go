package plugin

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/containers/common/pkg/hooks"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/opea-project/GenAIInfra/kubernetes-addons/memory-bandwidth-exporter/info"
	rspec "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/containerd/nri/pkg/api"
	"github.com/containerd/nri/pkg/stub"
)

const (
	NriPluginName  = "memory-bandwidth-exporter.v1"
	conSuffix      = ".scope"
	podPrefix      = "/sys/fs/cgroup"
	conPrefix      = "/cri-containerd-"
	rootMonDir     = "/sys/fs/resctrl/mon_groups/"
	monGroupPrefix = "container-"
	annRDT         = "rdt.resources.beta.kubernetes.io/pod"
)

var (
	isNeedMakeMonitorGroup = true
)

type Plugin struct {
	PluginName   string
	PluginIdx    string
	DisableWatch bool
	Stub         stub.Stub
	Mgr          *hooks.Manager
	Logger       log.Logger
}

func (p *Plugin) Synchronize(_ context.Context, pod []*api.PodSandbox, container []*api.Container) ([]*api.ContainerUpdate, error) {
	dropRepeat := make(map[string]int)
	for _, container := range container {
		if _, ok := dropRepeat[container.Id]; ok {
			continue
		}
		dropRepeat[container.Id] = 1
		for _, pod := range pod {
			if _, ok := pod.Annotations[annRDT]; ok {
				continue
			}
			if pod.Id == container.PodSandboxId {
				cgroupPath := podPrefix + pod.Linux.CgroupParent + conPrefix + container.Id + conSuffix
				monPath := rootMonDir + monGroupPrefix + container.Id
				cif := info.ContainerInfo{
					Operation:     2,
					ContainerName: container.Name,
					ContainerId:   container.Id,
					PodName:       pod.Name,
					NameSpace:     pod.Namespace,
					CgroupPath:    cgroupPath,
					MonGroupPath:  monPath,
				}
				ContainerInfoes := make(map[string]info.ContainerInfo)
				ContainerInfoes[container.Id] = cif
				if len(info.ContainerInfoChan) == cap(info.ContainerInfoChan) {
					return nil, fmt.Errorf("ContainerInfoChan is full")
				}
				info.ContainerInfoChan <- ContainerInfoes
				break
			}
		}
	}
	return nil, nil
}

func (p *Plugin) CreateContainer(_ context.Context, pod *api.PodSandbox, container *api.Container) (*api.ContainerAdjustment, []*api.ContainerUpdate, error) {
	if !isNeedMakeMonitorGroup {
		return nil, nil, nil
	}

	ctrName := containerName(pod, container)

	if val, ok := pod.Annotations[annRDT]; ok {
		level.Info(p.Logger).Log("msg", "container %v has rdt annotation %v", ctrName, val)
		return nil, nil, nil
	}

	annotations := map[string]string{}
	for k, v := range container.Annotations {
		annotations[k] = v
	}
	for k, v := range pod.Annotations {
		annotations[k] = v
	}
	hasBindMounts := len(container.Mounts) > 0

	spec := &rspec.Spec{
		Process: &rspec.Process{
			Args: container.Args,
		},
	}

	if _, err := p.Mgr.Hooks(spec, annotations, hasBindMounts); err != nil {
		level.Error(p.Logger).Log("msg", "failed to generate hooks", "container", ctrName, "err", err)
		return nil, nil, fmt.Errorf("hook generation failed: %w", err)
	}

	if spec.Hooks == nil {
		level.Info(p.Logger).Log("msg", "container %v has no hooks to inject, ignoring", ctrName)
		return nil, nil, nil
	}

	adjust := &api.ContainerAdjustment{}
	adjust.AddHooks(api.FromOCIHooks(spec.Hooks))
	level.Info(p.Logger).Log("msg", "OCI hooks injected", "container", ctrName)

	return adjust, nil, nil
}

func (p *Plugin) StartContainer(_ context.Context, pod *api.PodSandbox, container *api.Container) error {
	if _, ok := pod.Annotations[annRDT]; ok {
		return nil
	}

	level.Info(p.Logger).Log("msg", "StartContainer stage", "container frist pid", container.Pid)

	cif := info.ContainerInfo{
		Operation:     1,
		ContainerName: container.Name,
		ContainerId:   container.Id,
		PodName:       pod.Name,
		NameSpace:     pod.Namespace,
		CgroupPath:    podPrefix + pod.Linux.CgroupParent + conPrefix + container.Id + conSuffix,
		MonGroupPath:  rootMonDir + monGroupPrefix + container.Id,
	}
	ContainerInfoes := make(map[string]info.ContainerInfo)
	ContainerInfoes[container.Id] = cif
	if len(info.ContainerInfoChan) == cap(info.ContainerInfoChan) {
		return fmt.Errorf("ContainerInfoChan is full")
	}
	info.ContainerInfoChan <- ContainerInfoes
	return nil
}

func (p *Plugin) StopContainer(_ context.Context, pod *api.PodSandbox, container *api.Container) ([]*api.ContainerUpdate, error) {
	if _, ok := pod.Annotations[annRDT]; ok {
		return nil, nil
	}

	cif := info.ContainerInfo{
		Operation:     0,
		ContainerName: container.Name,
		ContainerId:   container.Id,
	}
	ContainerInfoes := make(map[string]info.ContainerInfo)
	ContainerInfoes[container.Id] = cif
	if len(info.ContainerInfoChan) == cap(info.ContainerInfoChan) {
		return nil, fmt.Errorf("ContainerInfoChan is full")
	}
	info.ContainerInfoChan <- ContainerInfoes
	return nil, nil
}

func containerName(pod *api.PodSandbox, container *api.Container) string {
	if pod != nil {
		return pod.Name + "/" + container.Name
	}
	return container.Name
}

func (p *Plugin) Run(isNeed bool) error {
	isNeedMakeMonitorGroup = isNeed
	var (
		opts []stub.Option
		mgr  *hooks.Manager
		err  error
	)

	if p.PluginName != "" {
		opts = append(opts, stub.WithPluginName(p.PluginName))
	}
	if p.PluginIdx != "" {
		opts = append(opts, stub.WithPluginIdx(p.PluginIdx))
	}

	if p.Stub, err = stub.New(p, opts...); err != nil {
		return fmt.Errorf("failed to create plugin stub: %v", err)
	}

	ctx := context.Background()
	dirs := []string{hooks.DefaultDir, hooks.OverrideDir}
	mgr, err = hooks.New(ctx, dirs, []string{})
	if err != nil {
		return fmt.Errorf("failed to set up hook manager: %v", err)
	}
	p.Mgr = mgr

	if !p.DisableWatch {
		for _, dir := range dirs {
			if err = os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %q: %v", dir, err)
			}
		}

		sync := make(chan error, 2)
		go mgr.Monitor(ctx, sync)

		err = <-sync
		if err != nil {
			return fmt.Errorf("failed to monitor hook directories: %v", err)
		}
		level.Info(p.Logger).Log("msg", "watching directories for new changes", "dirs", strings.Join(dirs, " "))
	}

	err = p.Stub.Run(ctx)
	if err != nil {
		return fmt.Errorf("plugin exited with error %v", err)
	}
	return nil
}
