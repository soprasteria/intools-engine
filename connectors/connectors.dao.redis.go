package connectors

import (
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/soprasteria/intools-engine/executors"
	"github.com/soprasteria/intools-engine/intools"
)

func GetRedisConnectorsKey(c *Connector) string {
	return "intools:groups:" + c.Group + ":connectors"
}

func GetRedisConnectorKey(c *Connector) string {
	return GetRedisrKey(c.Group, c.Name)
}

func GetRedisrKey(g string, c string) string {
	return "intools:groups:" + g + ":connectors:" + c
}

func GetRedisConnectorConfKey(g string, c string) string {
	return GetRedisrKey(g, c) + ":conf"
}

func RedisGetConnectors(group string) ([]string, error) {
	r := intools.Engine.GetRedisClient()
	key := fmt.Sprintf("intools:groups:%s:connectors", group)
	len, err := r.LLen(key).Result()
	if err != nil {
		return nil, err
	}
	return r.LRange(key, 0, len).Result()
}

func RedisGetConnector(group string, connector string) (*Connector, error) {
	r := intools.Engine.GetRedisClient()
	log.Debugf("Loading %s:%s from redis", group, connector)
	key := GetRedisConnectorConfKey(group, connector)
	cmd := r.Get(key)
	jsonCmd := cmd.Val()
	if cmd.Err() != nil {
		log.WithError(cmd.Err()).Error("Redis command failed")
		return nil, errors.New("Unable to load connectors " + group + "/" + connector + ":" + cmd.Err().Error())
	}
	c := &Connector{}
	err := json.Unmarshal([]byte(jsonCmd), c)
	if err != nil {
		log.Error("JSON Unmarshall failed with following value")
		log.Error(jsonCmd)
		return nil, err
	}
	return c, nil
}

func RedisSaveConnector(c *Connector) error {
	r := intools.Engine.GetRedisClient()
	log.Debugf("Saving %s to redis", c.Group)
	multi := r.Multi()
	defer multi.Close()
	_, err := multi.Exec(func() error {
		multi.LRem(GetRedisrKey(c.Group, c.Name), 0, c.Group)
		multi.LPush(GetRedisrKey(c.Group, c.Name), c.Group)
		multi.LRem(GetRedisConnectorsKey(c), 0, c.Name)
		multi.LPush(GetRedisConnectorsKey(c), c.Name)
		multi.Set(GetRedisConnectorConfKey(c.Group, c.Name), c.GetJSON(), 0)
		return nil
	})
	return err
}

func RedisRemoveConnector(c *Connector) error {
	r := intools.Engine.GetRedisClient()
	log.Debugf("Removing %s:%s from redis", c.Group, c.Name)
	multi := r.Multi()
	defer multi.Close()
	_, err := multi.Exec(func() error {
		multi.Del(GetRedisConnectorConfKey(c.Group, c.Name))
		multi.Del(GetRedisExecutorKey(c))
		multi.Del(GetRedisResultKey(c))
		multi.Del(GetRedisConnectorsKey(c))
		multi.Del(GetRedisConnectorsKey(c))
		multi.Del(GetRedisrKey(c.Group, c.Name))
		multi.Del(GetRedisrKey(c.Group, c.Name))
		return nil
	})
	return err
}

func GetRedisExecutorKey(c *Connector) string {
	return "intools:groups:" + c.Group + ":connectors:" + c.Name + ":executors"
}

func GetRedisResultKey(c *Connector) string {
	return "intools:groups:" + c.Group + ":connectors:" + c.Name + ":results"
}

func RedisSaveExecutor(c *Connector, exec *executors.Executor) error {
	r := intools.Engine.GetRedisClient()
	log.WithField("containerName", c.GetContainerName()).WithField("containerId", exec.ContainerId).Debug("Saving execution of connector to Redis")
	cmd := r.Set(GetRedisExecutorKey(c), exec.GetJSON(), 0)
	if exec.Valid {
		_ = r.Set(GetRedisResultKey(c), exec.GetResult(), 0)
	}
	return cmd.Err()
}

func RedisGetLastExecutor(c *Connector) (string, error) {
	r := intools.Engine.GetRedisClient()
	cmd := r.Get(GetRedisExecutorKey(c))
	if cmd.Err() != nil {
		return "", cmd.Err()
	}
	return cmd.Val(), nil
}
