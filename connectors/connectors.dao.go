package connectors

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/soprasteria/intools-engine/executors"
)

func SaveExecutor(c *Connector, exec *executors.Executor) {
	err := RedisSaveExecutor(c, exec)
	if err != nil {
		log.WithError(err).Error("Error while saving to Redis")
	}
}

func SaveConnector(c *Connector) {
	err := RedisSaveConnector(c)
	if err != nil {
		log.WithError(err).Error("Error while saving to Redis")
	}
}

func RemoveConnector(c *Connector) {
	err := RedisRemoveConnector(c)
	if err != nil {
		log.WithError(err).Error("Error while removing from Redis")
	}
}

func GetLastConnectorExecutor(c *Connector) *executors.Executor {
	sExecutor, err := RedisGetLastExecutor(c)
	if err != nil {
		log.WithError(err).Errorf("Cannot load last executor %s:%s from Redis", c.Group, c.Name)
		return nil
	}
	executor := &executors.Executor{}
	err = json.Unmarshal([]byte(sExecutor), executor)
	if err != nil {
		log.WithError(err).Errorf("Cannot parse last executor %s:%s", c.Group, c.Name)
		log.Error(err.Error())
		return nil
	}
	return executor
}

func GetConnector(group string, connector string) (*Connector, error) {
	conn, err := RedisGetConnector(group, connector)
	if err != nil {
		log.WithError(err).Errorf("Error while loading %s:%s to Redis", group, connector)
		return nil, err
	}
	return conn, nil
}

func GetConnectors(group string) []Connector {
	ret, err := RedisGetConnectors(group)
	if err != nil {
		log.WithError(err).Errorf("Error while getting connectors for group %s from Redis", group)
		return nil
	}
	connectors := make([]Connector, len(ret))
	for i, c := range ret {
		conn, err := GetConnector(group, c)
		if err != nil {
			log.Warnf("Unable to load %s:%s", group, c)
		} else {
			connectors[i] = *conn
		}
	}
	return connectors
}
