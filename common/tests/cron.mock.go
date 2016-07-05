package tests

import (
	"github.com/stretchr/testify/mock"
	"gopkg.in/robfig/cron.v2"
)

type CronMock struct {
	mock.Mock
	jobs map[string]cron.Job
}

func (c *CronMock) AddJob(spec string, cmd cron.Job) (cron.EntryID, error) {
	args := c.Called(spec, cmd)
	var entryID cron.EntryID
	return entryID, args.Error(0)
}

func (c *CronMock) Remove(id cron.EntryID) {
	c.Called(id)
}

func (c *CronMock) Schedule(schedule cron.Schedule, cmd cron.Job) cron.EntryID {
	c.Called(schedule, cmd)
	var entryID cron.EntryID
	return entryID
}

func (c *CronMock) Entries() []cron.Entry {
	return c.entrySnapshot()
}

func (c *CronMock) Start() {
	c.Called()
}

func (c *CronMock) run() {
	c.Called()
}

func (c *CronMock) Stop() {
	c.Called()
}

func (c *CronMock) entrySnapshot() []cron.Entry {
	c.Called()
	entries := []cron.Entry{}
	return entries
}
