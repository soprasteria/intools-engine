package groups

import (
	log "github.com/Sirupsen/logrus"
	"github.com/soprasteria/intools-engine/connectors"
)

func GetGroup(name string, withConnectors bool) *Group {
	allGroups := GetGroups(withConnectors)
	for _, g := range allGroups {
		if g.Name == name {
			return &g
		}
	}
	return nil
}

func GetGroupsLength() int64 {
	length, err := RedisGetLength()
	if err != nil {
		log.Errorf("Error while getting groups length from Redis %s", err.Error())
		return 0
	}
	return length
}

func GetGroups(withConnectors bool) []Group {
	groups, err := RedisGetGroups()
	if err != nil {
		log.Errorf("Error while getting groups from Redis %s", err.Error())
		return nil
	}
	allGroups := make([]Group, len(groups))
	for i, g := range groups {
		group := &Group{
			Name: g,
		}
		if withConnectors {
			connectors := connectors.GetConnectors(g)
			group.Connectors = connectors
		}
		allGroups[i] = *group
	}
	return allGroups
}

func CreateGroup(group string) (bool, error) {
	return RedisCreateGroup(group)
}

func DeleteGroup(group string) error {
	return RedisDeleteGroup(group)
}
