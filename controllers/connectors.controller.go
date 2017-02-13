package controllers

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/soprasteria/intools-engine/connectors"
)

func ControllerGetConnectors(c *gin.Context) {
	group := c.Param("group")
	connectors := connectors.GetConnectors(group)
	c.JSON(http.StatusOK, connectors)
}

func ControllerGetConnector(c *gin.Context) {
	group := c.Param("group")
	connector := c.Param("connector")

	log.Debugf("Searching for %s:%s", group, connector)

	conn, err := connectors.GetConnector(group, connector)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
	} else {
		c.JSON(http.StatusOK, conn)
	}
}

func ControllerExecConnector(c *gin.Context) {
	group := c.Param("group")
	connector := c.Param("connector")

	log.Debugf("Searching for %s:%s", group, connector)

	conn, err := connectors.GetConnector(group, connector)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
	} else {
		executor, err := connectors.Exec(conn)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.JSON(http.StatusOK, executor)
		}
	}
}

func ControllerGetConnectorExecutor(c *gin.Context) {
	group := c.Param("group")
	connector := c.Param("connector")
	conn, err := connectors.GetConnector(group, connector)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
	} else {
		exec := connectors.GetLastConnectorExecutor(conn)
		if exec == nil {
			c.String(http.StatusNotFound, "no executor found")
		} else {
			c.JSON(http.StatusOK, exec)
		}
	}
}

func ControllerGetConnectorResult(c *gin.Context) {
	group := c.Param("group")
	connector := c.Param("connector")
	conn, err := connectors.GetConnector(group, connector)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
	} else {
		exec := connectors.GetLastConnectorExecutor(conn)
		if exec == nil {
			c.String(http.StatusNotFound, "no result found")
		} else {
			if exec.Valid {
				c.JSON(http.StatusOK, exec.JsonStdout)
			} else {
				c.String(http.StatusNotFound, "invalid result")
			}
		}
	}
}

func ControllerCreateConnector(c *gin.Context) {
	group := c.Param("group")
	connector := c.Param("connector")

	var conn connectors.Connector
	c.Bind(&conn)
	conn.Group = group
	conn.Name = connector

	// Save Connector into Redis
	connectors.SaveConnector(&conn)

	// Schedule further executions
	connectors.Scheduler.SetJob(&conn)

	// Execute the connector
	go connectors.Exec(&conn)

	c.JSON(http.StatusOK, conn)
}

func ControllerDeleteConnector(c *gin.Context) {
	group := c.Param("group")
	connector := c.Param("connector")

	conn, err := connectors.GetConnector(group, connector)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
	}

	// Remove further scheduled executions
	connectors.Scheduler.RemoveJob(conn)

	// Remove connector
	connectors.RemoveConnector(conn)

	c.JSON(http.StatusOK, conn)
}
