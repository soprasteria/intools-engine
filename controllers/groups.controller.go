package controllers

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/soprasteria/intools-engine/connectors"
	"github.com/soprasteria/intools-engine/groups"
)

func ControllerGetGroups(c *gin.Context) {
	groups := groups.GetGroups(false)
	c.JSON(http.StatusOK, groups)
}

func ControllerGetGroup(c *gin.Context) {
	group := c.Param("group")
	g := groups.GetGroup(group, false)
	if g == nil {
		c.String(http.StatusNotFound, "")
	} else {
		c.JSON(http.StatusOK, g)
	}
}

func ControllerPostGroup(c *gin.Context) {
	group := c.Param("group")
	created, err := groups.CreateGroup(group)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	} else {
		if created {
			c.String(http.StatusCreated, "%s created", group)
		} else {
			c.String(http.StatusOK, "%s already exists", group)
		}
	}
}

func ControllerDeleteGroup(c *gin.Context) {
	group := c.Param("group")
	err := groups.DeleteGroup(group)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}

	conns, err := connectors.RedisGetConnectors(group)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}

	for _, cId := range conns {
		conn, err := connectors.GetConnector(group, cId)
		if err != nil {
			log.WithError(err).WithField("connectorId", cId).Error("Cant get connector from Redis")
		}
		// Remove further scheduled executions
		connectors.Scheduler.RemoveJob(conn)
		connectors.RemoveConnector(conn)
	}

	c.String(http.StatusOK, "%s deleted", group)

}
