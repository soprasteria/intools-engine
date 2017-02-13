package connectors

import (
	"encoding/json"
	"math/rand"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/orcaman/concurrent-map"
	"github.com/soprasteria/dockerapi"
)

var Scheduler ConnectorScheduler

func init() {
	Scheduler = NewConnectorScheduler()
}

type Connector struct {
	Group           string                      `json:"group"`
	Name            string                      `json:"name"`
	ContainerConfig *dockerapi.ContainerOptions `json:"config"`
	Timeout         uint                        `json:"timeout,omitempty"`
	Refresh         uint                        `json:"refresh,omitempty"`
}

type ConnectorScheduler struct {
	connectorTickers cmap.ConcurrentMap
}

func NewConnectorScheduler() ConnectorScheduler {
	return ConnectorScheduler{connectorTickers: cmap.New()}
}

func (ct ConnectorScheduler) SetJob(conn *Connector) {

	log.WithField("Group", conn.Group).WithField("Name", conn.Name).Info("Setting scheduling of connector...")

	if tmp, ok := ct.connectorTickers.Get(conn.Id()); ok {
		oldTicker := tmp.(*time.Ticker)
		oldTicker.Stop()
		ct.connectorTickers.Remove(conn.Id())
		log.WithField("Group", conn.Group).WithField("Name", conn.Name).Info("Stopped old scheduling job for connector")
	}

	newTicker := ct.newTicker(conn)
	ct.connectorTickers.Set(conn.Id(), newTicker)

	log.WithFields(log.Fields{
		"Group":              conn.Group,
		"Name":               conn.Name,
		"Refresh in minutes": conn.Refresh,
	}).Info("Connector is scheduled")

	log.Infof("There are %v connectors now scheduled", ct.connectorTickers.Count())
}

func (ct ConnectorScheduler) RemoveJob(conn *Connector) {

	log.WithField("Group", conn.Group).WithField("Name", conn.Name).Info("Removing scheduling of connector...")

	if tmp, ok := ct.connectorTickers.Get(conn.Id()); ok {
		oldTicker := tmp.(*time.Ticker)
		ct.connectorTickers.Remove(conn.Id())
		oldTicker.Stop()
		log.WithField("Group", conn.Group).WithField("Name", conn.Name).Info("Stopped old scheduling job for connector")
	} else {
		log.WithField("Group", conn.Group).WithField("Name", conn.Name).Warn("Unable to remove unexisting job")
	}

	log.WithField("Group", conn.Group).WithField("Name", conn.Name).Info("Connector is not scheduled anymore")
	log.Infof("There are %v connectors now scheduled", ct.connectorTickers.Count())
}

// getRandomizedRefreshTime generates a duration depending on following rules :
// - From refreshInMinutes, get a random duration around -2m and +2m -> duration-2m < effective duration < duration+2m
// - If effective duration is under 1m, set a default random duration between 1m and 5m
func getRandomizedRefreshTime(refreshInMinutes uint) time.Duration {
	randomDuration := time.Duration(rand.Intn(240)-120) * time.Second // Random duration between -2min and +2min
	duration := time.Duration(refreshInMinutes)*time.Minute + randomDuration
	if duration <= time.Duration(1*time.Minute) {
		duration = time.Duration(rand.Intn(4)+1) * time.Minute
	}
	return duration
}

func (ct ConnectorScheduler) newTicker(conn *Connector) *time.Ticker {

	duration := getRandomizedRefreshTime(conn.Refresh)

	ticker := time.NewTicker(duration)
	log.WithFields(log.Fields{
		"Group":              conn.Group,
		"Name":               conn.Name,
		"Refresh in minutes": conn.Refresh,
	}).Infof("Ticker will next be executed in %v", duration.String())

	go func() {
		for _ = range ticker.C {
			Exec(conn)
			log.WithField("Group", conn.Group).WithField("Name", conn.Name).Infof("Connector executed. Next execution in %s", duration.String())
		}
	}()

	return ticker
}

func NewConnector(group string, name string) *Connector {
	conn := &Connector{group, name, nil, 15, 300}
	return conn
}

func (c *Connector) Init(image string, timeout uint, refresh uint, cmd []string) {
	if c.ContainerConfig == nil {
		c.ContainerConfig = &dockerapi.ContainerOptions{
			Image: image,
			Cmd:   cmd,
			Name:  c.Name,
		}
	}

	if timeout != 0 {
		c.Timeout = timeout
	}
	if refresh != 0 {
		c.Refresh = refresh
	}
}

func (c *Connector) GetContainerName() string {
	return c.ContainerConfig.Name
}

func (c *Connector) GetJSON() string {
	b, err := json.Marshal(c)
	if err != nil {
		log.Error(err)
		return ""
	}
	return string(b[:])
}

func (c *Connector) Id() string {
	return c.Group + ":" + c.Name
}

func (c *Connector) Run() {
	//TODO : Should not run error, or invalid connector ?
	log.Debug("Run Connector %s:%s", c.Group, c.Name)
	Exec(c)
}
