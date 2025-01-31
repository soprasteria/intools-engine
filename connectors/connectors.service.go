package connectors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/soprasteria/intools-engine/common/websocket"
	"github.com/soprasteria/intools-engine/executors"
	"github.com/soprasteria/intools-engine/intools"
)

func Exec(connector *Connector) (*executors.Executor, error) {
	executor := &executors.Executor{}

	//Saving connector to redis
	go SaveConnector(connector)

	//Get all containers
	simpleContainers, err := intools.Engine.GetDockerClient().ListContainers()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	//Searching for the container with the same name
	containerExists := false
	previousContainerID := "-1"
	for _, c := range simpleContainers.GetAll() {
		if c.Name() == fmt.Sprintf("/%s", connector.GetContainerName()) {
			containerExists = true
			previousContainerID = c.ID()
		}
	}

	//If it exists, remove it
	if containerExists {
		log.Infof("Removing container %s [/%s]", previousContainerID[:11], connector.GetContainerName())
		removeContainerOptions := docker.RemoveContainerOptions{ID: previousContainerID, RemoveVolumes: true, Force: true}
		err = intools.Engine.GetDockerClient().Docker.RemoveContainer(removeContainerOptions)
		if err != nil {
			log.Error("Cannot remove container " + previousContainerID[:11])
			log.Error(err)
			return nil, err
		}
	}

	//Create container
	log.Debug("New container with config ", connector.ContainerConfig)
	container, err := intools.Engine.GetDockerClient().NewContainer(*connector.ContainerConfig)
	if err != nil {
		log.Error("Cannot create container " + connector.GetContainerName())
		log.Error(err)
		return nil, err
	}
	//Save the short ContainerId
	executor.Host = intools.Engine.GetDockerHost()

	// Starting container
	log.Info("Starting container " + connector.GetContainerName())
	// we don't want to force pulling of images in order to support projects which don't have a registry because images are only local in that case
	forcePull := false
	err = container.Run(forcePull)
	if err != nil {
		log.Error("Cannot start container " + connector.GetContainerName())
		log.Error(err)
		return nil, err
	}

	//Prepare the waiting group to sync execution of the container
	var wg sync.WaitGroup
	wg.Add(1)

	executor.ContainerId = container.ID()[:11]
	log.WithField("containerId", executor.ContainerId).WithField("containerName", connector.GetContainerName()).Info("Container successfully started")
	log.Debug(executor.ContainerId + " will be stopped after " + fmt.Sprint(connector.Timeout) + " seconds")
	//Trigger stop of the container after the timeout
	intools.Engine.GetDockerClient().Docker.StopContainer(container.ID(), connector.Timeout)

	//Wait for the end of the execution of the container
	for {
		//Each time inspect the container
		inspect, err := intools.Engine.GetDockerClient().InspectContainer(container.ID())
		if err != nil {
			log.Error("Cannot inspect container " + connector.GetContainerName())
			log.Error(err)
			return executor, err
		}
		if !inspect.IsRunning() {
			//When the container is not running
			log.Debug(connector.GetContainerName() + " is stopped")
			executor.Running = false
			executor.Terminated = true
			executor.ExitCode = inspect.Container.State.ExitCode
			executor.StartedAt = inspect.Container.State.StartedAt
			executor.FinishedAt = inspect.Container.State.FinishedAt
			//Trigger next part of the waiting group
			wg.Done()
			//Exit from the waiting loop
			break
		} else {
			//Wait
			log.Debug(connector.GetContainerName() + " is running...")
			time.Sleep(5 * time.Second)
		}
	}

	//Next part : after the container has been executed
	wg.Wait()

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	logOptions := docker.LogsOptions{
		Container:    container.ID(),
		OutputStream: stdoutBuf,
		ErrorStream:  stderrBuf,
		Stdout:       true,
		Stderr:       true,
		Tail:         "all",
		Follow:       true,
		Timestamps:   false,
	}

	//Get the stdout and stderr
	err = intools.Engine.GetDockerClient().Docker.Logs(logOptions)

	if err != nil {
		log.Error("-cannot read stdout logs from server")
	} else {
		containerLogs := stdoutBuf.String()
		log.Debugf("container logs %s", containerLogs)
		executor.Stdout = containerLogs
		executor.JsonStdout = new(map[string]interface{})
		errJSONStdOut := json.Unmarshal(stdoutBuf.Bytes(), executor.JsonStdout)
		executor.Valid = true

		if errJSONStdOut != nil {
			log.Warnf("Unable to parse stdout from container %s", container.Name())
			log.Warnf("Error: %s - Stdout: %s", errJSONStdOut, containerLogs)
		}

		executor.Stderr = stderrBuf.String()
	}

	removeVolumes := false
	err = container.Remove(removeVolumes)
	if err != nil {
		log.Error("Cannot remove container " + container.Name())
		log.Error(err)
		return nil, err
	}

	// Broadcast result to registered clients
	lightConnector := &websocket.LightConnector{
		GroupId:     connector.Group,
		ConnectorId: connector.Name,
		Value:       executor.JsonStdout,
	}
	websocket.ConnectorBuffer <- lightConnector

	//Save result to redis
	defer SaveExecutor(connector, executor)

	return executor, nil
}
