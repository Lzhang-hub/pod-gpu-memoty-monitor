# gpu-memory-monitor

本项目用于监控k8s集群上pod中进程使用的GPU显存大小，主要是针对共享GPU的情况。

如果是单卡的情况可以参看：https://github.com/king-jingxiang/pod-gpushare-metrics-exporter，该方案可以直接监控每张卡上面的pod的资源使用情况。

主要原理：

- 调用nvml库获取每张卡上面占用GPU的进程信息，该信息包含了进程的显存占用；
- 调用docker.sock获取当前机器上面所有docker容器信息，主要包含容器的PID、容器对应的pod信息；
- 将卡上的进程信息和docker中的进程信息匹配，如果存在，则返回该容器对应的pod信息，（注意：GPU卡上的进程可能是容器进程的子进程，所以需要判断GPU卡上进程的父进程在容器列表中是否存在）

### Prerequisites
- golang 1.15+
- NVIDIA drivers ~= 361.93
- Nvidia-docker version > 2.0 (see how to [install](https://github.com/NVIDIA/nvidia-docker) and it's [prerequisites](https://github.com/nvidia/nvidia-docker/wiki/Installation-\(version-2.0\)#prerequisites))

### How to build binary?
```
$ git clone https://github.com/lxyzhangqing/gpu-memory-monitor.git
$ cd gpu-memory-monitor
$ go mod tidy
$ go mod vendor
$ make
```

### How to build images?
```
$ git clone https://github.com/lxyzhangqing/gpu-memory-monitor.git
$ cd gpu-memory-monitor
$ go mod tidy
$ go mod vendor
$ docker build -t gpu-memory-monitor:v1 .
```

### How to deploy gpu memory monitor by docker?
You can execute the following command line on your GPU machine.
```
docker run -d --name=gpu-memory-monitor -e NVIDIA_VISIBLE_DEVICES=all -e NVIDIA_DRIVER_CAPABILITIES=utility -v /var/run:/var/run:ro  --net=host gpu-memory-monitor:v1
```

### How to deploy gpu memory monitor by kubernetes?
You can copy `deploy.yaml` to your kubernetes cluster and execute the following command line to deploy `gpu-momory-monitor`. 
Before this, you should to edit `nodeAffinity` for scheduling pods of `gpu-memory-monitor` metrics server to correct GPU machines.
```
kubectl create -f deploy.yaml
```

### How to get the metrics?
You can execute this command line on you machine:
```
curl http://127.0.0.1:5091/metrics
```

Then you may get metrics info like this:
```
# HELP pod gpu memory usage, unit is MiB
# TYPE pod_gpu_memory_usage gauge
pod_gpu_memory_usage{gpu_type="Tesla T4",gpu_uuid="GPU-576ab88b-464f-5903-3ab9-2d25e3ee6c4a",hostname="test-node",name="gpu.test1-85846f7bd4-4ppm9",namespace="default",pid="37691"} 2027
pod_gpu_memory_usage{gpu_type="Tesla T4",gpu_uuid="GPU-6758250c-1793-6349-ba37-332ac77b1d0a",hostname="test-node",name="gpu.test2-57485d95d6-wsngh",namespace="default",pid="54702"} 3449
```