package tests

import (
	"github.com/soprasteria/dockerapi"
	"github.com/soprasteria/intools-engine/intools"
)

type IntoolsEngineMock struct {
	DockerClient dockerapi.Client
	DockerHost   string
	RedisClient  intools.RedisWrapper
	Cron         intools.CronWrapper
}

func (e IntoolsEngineMock) GetDockerClient() *dockerapi.Client {
	return &e.DockerClient
}

func (e IntoolsEngineMock) GetDockerHost() string {
	return e.DockerHost
}

func (e IntoolsEngineMock) GetRedisClient() intools.RedisWrapper {
	return e.RedisClient
}

func (e IntoolsEngineMock) GetCron() intools.CronWrapper {
	return e.GetCron()
}
