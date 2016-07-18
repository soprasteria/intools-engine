package groups_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/soprasteria/dockerapi"
	"github.com/soprasteria/intools-engine/common/tests"

	"github.com/soprasteria/intools-engine/groups"
	"github.com/soprasteria/intools-engine/intools"
)

var _ = Describe("Groups", func() {

	var (
		engine tests.IntoolsEngineMock
		cron   tests.CronMock
		redis  intools.RedisWrapper
		docker dockerapi.Client
	)

	BeforeEach(func() {
		cron = tests.CronMock{}
		redis = &tests.RedisClientMock{}
		docker = dockerapi.Client{}
		engine = tests.IntoolsEngineMock{DockerClient: docker, DockerHost: "mock.local:2576", RedisClient: redis, Cron: &cron}

		intools.Engine = engine
	})

	Describe("Reloading Data from Redis Store", func() {
		Context("With no Redis Store", func() {
			It("Should do nothing", func() {
				groups.Reload()
				Expect(cron.AssertNumberOfCalls(GinkgoT(), "AddJob", 0))
			})
		})
	})
})
