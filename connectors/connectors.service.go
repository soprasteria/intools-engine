package connectors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/soprasteria/intools-engine/common/logs"
	"github.com/soprasteria/intools-engine/common/websocket"
	"github.com/soprasteria/intools-engine/executors"
	"github.com/soprasteria/intools-engine/intools"
	"gopkg.in/robfig/cron.v2"
)

func InitSchedule(c *Connector) cron.EntryID {
	if intools.Engine.GetCron() != nil {
		crontab := fmt.Sprintf("@every %dm", c.Refresh)
		logs.Debug.Printf("Schedule %s:%s %s", c.Group, c.Name, crontab)
		entryId, _ := intools.Engine.GetCron().AddJob(crontab, c)
		return entryId
	}
	return 0
}

func RemoveScheduleJob(entryId cron.EntryID) {
	if intools.Engine.GetCron() != nil {
		logs.Debug.Printf("Remove schedule job with cronId: %s", entryId)
		intools.Engine.GetCron().Remove(entryId)
	}
}

func Exec(connector *Connector) (*executors.Executor, error) {
	executor := &executors.Executor{}

	//Saving connector to redis
	go SaveConnector(connector)

	//Get all containers
	containers, err := intools.Engine.GetDockerClient().ListContainers()
	if err != nil {
		logs.Error.Println(err)
		return nil, err
	}

	//Searching for the container with the same name
	containerExists := false
	previousContainerID := "-1"
	for _, c := range containers {
		for _, n := range c.Container.Names {
			if n == fmt.Sprintf("/%s", connector.GetContainerName()) {
				containerExists = true
				previousContainerID = c.ID()
			}
		}
	}

	//If it exists, remove it
	if containerExists {
		logs.Info.Printf("Removing container %s [/%s]", previousContainerID[:11], connector.GetContainerName())
		removeContainerOptions := docker.RemoveContainerOptions{ID: previousContainerID, RemoveVolumes: true, Force: true}
		err = intools.Engine.GetDockerClient().Docker.RemoveContainer(removeContainerOptions)
		if err != nil {
			logs.Error.Println("Cannot remove container " + previousContainerID[:11])
			logs.Error.Println(err)
			return nil, err
		}
	}

	//Create container
	logs.Debug.Println("New container with config ", connector.ContainerConfig)
	container, err := intools.Engine.GetDockerClient().NewContainer(*connector.ContainerConfig)
	if err != nil {
		logs.Error.Println("Cannot create container " + connector.GetContainerName())
		logs.Error.Println(err)
		return nil, err
	}
	//Save the short ContainerId
	executor.Host = intools.Engine.GetDockerHost()

	// Starting container
	logs.Info.Println("Starting container " + connector.GetContainerName())
	err = container.Run()
	if err != nil {
		logs.Error.Println("Cannot start container " + connector.GetContainerName())
		logs.Error.Println(err)
		return nil, err
	}

	//Prepare the waiting group to sync execution of the container
	var wg sync.WaitGroup
	wg.Add(1)

	executor.ContainerId = container.ID()[:11]
	logs.Info.Printf("%s [/%s] successfully started", executor.ContainerId, connector.GetContainerName())
	logs.Debug.Println(executor.ContainerId + " will be stopped after " + fmt.Sprint(connector.Timeout) + " seconds")
	//Trigger stop of the container after the timeout
	intools.Engine.GetDockerClient().Docker.StopContainer(container.ID(), connector.Timeout)

	//Wait for the end of the execution of the container
	for {
		//Each time inspect the container
		inspect, err := intools.Engine.GetDockerClient().InspectContainer(container.ID())
		if err != nil {
			logs.Error.Println("Cannot inspect container " + connector.GetContainerName())
			logs.Error.Println(err)
			return executor, err
		}
		if !inspect.IsRunning() {
			//When the container is not running
			logs.Debug.Println(connector.GetContainerName() + " is stopped")
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
			logs.Debug.Println(connector.GetContainerName() + " is running...")
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
		logs.Error.Println("-cannot read stdout logs from server")
	} else {
		containerLogs := stdoutBuf.String()
		logs.Debug.Printf("container logs %s", containerLogs)
		executor.Stdout = containerLogs
		executor.JsonStdout = new(map[string]interface{})
		errJSONStdOut := json.Unmarshal(stdoutBuf.Bytes(), executor.JsonStdout)
		executor.Valid = true

		if errJSONStdOut != nil {
			logs.Warning.Printf("Unable to parse stdout from container %s", container.Name())
			logs.Warning.Printf("Error: %s - Stdout: %s", errJSONStdOut, containerLogs)
		}

		executor.Stderr = stderrBuf.String()
	}

	removeVolumes := false
	err = container.Remove(removeVolumes)
	if err != nil {
		logs.Error.Println("Cannot remove container " + container.Name())
		logs.Error.Println(err)
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
