package cli

import (
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"

	"github.com/soprasteria/intools-engine/common/server"
	"github.com/soprasteria/intools-engine/common/utils"
	"github.com/soprasteria/intools-engine/connectors"
	"github.com/soprasteria/intools-engine/groups"
	"github.com/soprasteria/intools-engine/intools"
)

func initLoggers(lvl string) {
	log.SetOutput(os.Stdout)
	level, err := log.ParseLevel(lvl)
	if err != nil {
		level = log.WarnLevel
		log.WithError(err).WithField("defaultLevel", level).Warn("Invalid log level, using default")
	}
	log.SetLevel(level)

	log.SetFormatter(&log.TextFormatter{})
}

func daemonAction(c *cli.Context) {
	port := c.GlobalInt("port")
	level := c.GlobalString("log-level")
	initLoggers(level)
	logPath := c.GlobalString("log-path")
	log.Info("Starting Intools-Engine as daemon")

	dockerClient, dockerHost, err := utils.GetDockerCient(c)
	if err != nil {
		os.Exit(1)
	}

	redisClient, err := utils.GetRedis(c)
	if err != nil {
		os.Exit(1)
	}

	d := server.NewDaemon(port, level, dockerClient, dockerHost, redisClient)
	d.SetRoutes(logPath)
	d.Run()
}

func runAction(c *cli.Context) {
	level := c.GlobalString("log-level")
	initLoggers(level)

	dockerClient, host, err := utils.GetDockerCient(c)
	if err != nil {
		os.Exit(1)
	}

	redisClient, err := utils.GetRedis(c)
	if err != nil {
		os.Exit(1)
	}

	cmd := []string{c.Args().First()}
	cmd = append(cmd, c.Args().Tail()...)
	if len(cmd) < 4 {
		log.Error("Incorrect usage, please check --help")
		return
	}
	group := cmd[0]
	conn := cmd[1]
	image := cmd[2]
	t := cmd[3]
	timeout, err := strconv.ParseUint(t, 10, 64)
	if err != nil {
		// handle error
		log.WithError(err).Error("Error while parsing timeout")
		os.Exit(2)
	}
	cmd = cmd[4:]

	log.WithFields(log.Fields{"image": image, "commands": cmd}).Debug("Launching...")
	log.Warn("In command line, connector schedule is not available")
	intools.Engine = &intools.IntoolsEngineImpl{DockerClient: dockerClient, DockerHost: host, RedisClient: redisClient, Cron: nil}
	connector := connectors.NewConnector(group, conn)
	connector.Init(image, uint(timeout), 0, cmd)
	groups.CreateGroup(group)
	if err != nil {
		os.Exit(3)
	}
	executor, err := connectors.Exec(connector)
	if err != nil {
		os.Exit(3)
	}
	log.Info(executor.GetJSON())

}

func testAction(c *cli.Context) {
	log.Error("Not yet implemented")
}

func publishAction(c *cli.Context) {
	log.Error("Not yet implemented")
}
