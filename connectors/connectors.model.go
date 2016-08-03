package connectors

import (
	"encoding/json"
	"strings"

	"github.com/soprasteria/dockerapi"
	"github.com/soprasteria/intools-engine/common/logs"
	"github.com/soprasteria/intools-engine/executors"
)

type Connector struct {
	Group           string                      `json:"group"`
	Name            string                      `json:"name"`
	ContainerConfig *dockerapi.ContainerOptions `json:"config"`
	Timeout         uint                        `json:"timeout,omitempty"`
	Refresh         uint                        `json:"refresh,omitempty"`
}

type ConnectorRunner interface {
	Exec(connector *Connector) (*executors.Executor, error)
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
		logs.Error.Println(err)
		return ""
	}
	return string(b[:])
}

func (c *Connector) Run() {
	//TODO : Should not run error, or invalid connector ?
	logs.Debug.Printf("Run Connector %s:%s", c.Group, c.Name)
	addQuotes(c.ContainerConfig)
	Exec(c)
}

func addQuotes(c *dockerapi.ContainerOptions) {
	var quotedCmd = make([]string, len(c.Cmd))
	for i, arg := range c.Cmd {
		if !strings.HasPrefix(arg, "\"") {
			quotedCmd[i] = "\"" + arg + "\""
		} else {
			quotedCmd[i] = arg
		}
	}
	c.Cmd = quotedCmd
	logs.Debug.Printf("Connector started with parameters = %s", quotedCmd)
}
