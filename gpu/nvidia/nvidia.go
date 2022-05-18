package nvidia

import (
	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
	"github.com/lxyzhangqing/gpu-memory-monitor/docker"
	"github.com/lxyzhangqing/gpu-memory-monitor/gpu/types"
	"log"
)

func Init() error {
	err := nvml.Init()
	if err != nil {
		log.Printf("Unable to initialize NVML: %v", err)
		return err
	}
	return nil
}

func Finish() {
	err := nvml.Shutdown()
	if err != nil {
		log.Printf("Unable to shutdown NVML: %v", err)
	}
}

func GetProcessInfo() ([]types.PodResourceUsage, error) {
	count, err := nvml.GetDeviceCount()
	if err != nil {
		log.Printf("Unable to get device count: %v", err)
		return nil, err
	}
	//获取容器列表，包含运行中容器的基本信息，进程PID，pod名称，pod的namespace
	containerList, err := docker.GetContainerList()
	if err != nil {
		log.Printf("get container list failed: %v", err)
		return nil, err
	}

	podsResourceUsage := make([]types.PodResourceUsage, 0)
	for i := uint(0); i < count; i++ {
		device, err := nvml.NewDevice(i)
		if err != nil {
			log.Printf("Unable to get device at index %d: %v", i, err)
			return nil, err
		}

		name := ""
		if device.Model != nil {
			name = *device.Model
		}
		uuid := device.UUID

		//获取GPU的进程PID
		pidArr, memoryArr, err := device.GetComputeRunningProcesses()
		if err != nil {
			log.Printf("Unabel to get processes info: %v", err)
			return nil, err
		}

		if len(pidArr) > 0 && len(pidArr) == len(memoryArr) {
			for i := 0; i < len(pidArr); i++ {
				//pidArr[i]代表GPU进程的PID,在容器列表总查看是否存在GPU进程的PID是否存在，如果存在返回对应容器的pod信息
				//存在bug，当GPU的进程的PID是容器进程PID的子进程时，返回结果为空，所以需要查找GPU进程的父进程在不在容器进程列表中
				podNamespace, podName, podHostName, check := containerList.GetGpuProcessPpid(uint32(pidArr[i]))
				//podNamespace, podName, podHostName := containerList.GetPodInfo(uint32(pidArr[i]))
				//
				if check {
					podsResourceUsage = append(podsResourceUsage, PodResourceUsage{
						Namespace: podNamespace,
						Name:      podName,
						HostName:  podHostName,
						PodPid:    uint32(pidArr[i]),
						GpuMemory: &GPU{
							DeviceName: name,
							DeviceUUID: uuid,
							MemoryUsed: memoryArr[i],
						},
					})
				}
			}
		}
	}

	return podsResourceUsage, nil
}
