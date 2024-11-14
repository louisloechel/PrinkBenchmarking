package prink

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"prinkbenchmarking/src/types"
	"strings"
	"text/template"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func StartPrink(experiment *types.Experiment, config types.Config) error {
	ctx := context.Background()

	// string writer
	writer := new(strings.Builder)
	tmpl, err := template.New("dockerHost").Parse(config.SutDockerHostTemplate)
	if err != nil {
		return err
	}

	err = tmpl.Execute(writer, map[string]string{
		"Address": experiment.SutHost,
	})
	if err != nil {
		return err
	}

	dockerHost := writer.String()

	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
		client.WithHost(dockerHost),
	)
	if err != nil {
		return err
	}
	defer cli.Close()

	// reader, err := cli.ImagePull(ctx, config.PrinkDockerImage, image.PullOptions{})
	if err := exec.Command("docker", "-H", dockerHost, "pull", config.PrinkDockerImage).Run(); err != nil {
		return err
	}

	networkName := "prink-eval" + experiment.ToFileName()
	net, err := cli.NetworkCreate(ctx, networkName, network.CreateOptions{})
	if err != nil {
		log.Print(err)
	}
	if net.ID != "" {
		defer cli.NetworkRemove(ctx, net.ID)
	}

	cmd := append([]string{"standalone-job"}, experiment.ToArgs()...)

	log.Printf("Starting prink with command:%v", cmd)

	exposedPorts := []string{"8081", "9249"}
	exposedPortsDocker := nat.PortSet{}
	for _, p := range exposedPorts {
		exposedPortsDocker[nat.Port(p+"/tcp")] = struct{}{}
	}

	portBindings := nat.PortMap{}
	for _, p := range exposedPorts {
		portBindings[nat.Port(p+"/tcp")] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: p,
			},
		}
	}

	containerJobManager, err := cli.ContainerCreate(ctx, &container.Config{
		Image:    config.PrinkDockerImage,
		Cmd:      cmd,
		Tty:      false,
		Hostname: "jobmanager",
		Env: []string{
			`FLINK_PROPERTIES=
     jobmanager.rpc.address: jobmanager
		 rest.profiling.enabled: true
		 rest.flamegraph.enabled: true
		 metrics.reporter.prom.factory.class: org.apache.flink.metrics.prometheus.PrometheusReporterFactory
		 metrics.reporter.prom.port: 9249`,
		},
		ExposedPorts: exposedPortsDocker,
	}, &container.HostConfig{
		PortBindings: portBindings,
	}, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"network": {
				NetworkID: networkName,
			},
		},
	}, nil, "")
	if err != nil {
		return err
	}

	defer cli.ContainerRemove(ctx, containerJobManager.ID, container.RemoveOptions{Force: true})

	containerTaskManager, err := cli.ContainerCreate(ctx, &container.Config{
		Image:    config.PrinkDockerImage,
		Cmd:      []string{"taskmanager"},
		Tty:      false,
		Hostname: "taskmanger",
		Env: []string{
			fmt.Sprintf(
				`FLINK_PROPERTIES=
     jobmanager.rpc.address: jobmanager
		 metrics.reporter.prom.factory.class: org.apache.flink.metrics.prometheus.PrometheusReporterFactory
		 metrics.reporter.prom.port: 9250
		 rest.profiling.enabled: true
		 rest.flamegraph.enabled: true
     taskmanager.numberOfTaskSlots: 1
     taskmanager.memory.process.size: %s`, config.TaskManagerMemory),
		},
		ExposedPorts: nat.PortSet{
			"9250/tcp": struct{}{},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			"9250/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "9250",
				},
			},
		},
	}, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"network": {
				NetworkID: networkName,
			},
		},
	}, nil, "")
	defer cli.ContainerRemove(ctx, containerTaskManager.ID, container.RemoveOptions{Force: true})

	if err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, containerTaskManager.ID, container.StartOptions{}); err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, containerJobManager.ID, container.StartOptions{}); err != nil {
		return err
	}

	statusCh, errCh := cli.ContainerWait(ctx, containerJobManager.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
	}

	logJobManager, err := cli.ContainerLogs(ctx, containerJobManager.ID, container.LogsOptions{ShowStdout: true})
	if err == nil {
		writeLogs(config.OutputFolder+"/flink_job_manager-"+time.Now().Format("2006-01-02.15:04:05")+experiment.ToFileName()+".log", &logJobManager)
	}

	logTaskManager, err := cli.ContainerLogs(ctx, containerTaskManager.ID, container.LogsOptions{ShowStdout: true})
	if err == nil {
		writeLogs(config.OutputFolder+"/flink_task_manager-"+time.Now().Format("2006-01-02.15:04:05")+experiment.ToFileName()+".log", &logTaskManager)
	}
	return nil
}

func writeLogs(filename string, src *io.ReadCloser) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Could not open logs file: %v", err)
	}
	defer file.Close()

	io.Copy(file, *src)
}
