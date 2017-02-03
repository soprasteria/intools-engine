package groups

import (
	log "github.com/Sirupsen/logrus"
	"github.com/soprasteria/intools-engine/connectors"
)

func Reload() {
	groups := GetGroups(true)
	for _, group := range groups {
		log.Infof("%s - Reloading group", group.Name)
		for _, connector := range group.Connectors {
			log.Infof("%s:%s - Reloading connector", group.Name, connector.Name)
			connectors.InitSchedule(&connector)
		}
	}
}
