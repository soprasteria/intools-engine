package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/soprasteria/intools-engine/common/utils"

	"github.com/gin-gonic/gin"
)

// GetLogs handler to retrieve logs
func GetLogs(c *gin.Context, path string) {
	content, err := getFileContent(path)
	format := c.Request.URL.Query().Get("format")
	if err == nil {
		if content == "" {
			err = errors.New("Logs not found in " + path)
			c.JSON(http.StatusNotFound, utils.HandleError("Logs not found", err, c))
		} else {
			switch format {
			case "text", "raw":
				c.String(http.StatusOK, content)
			default:
				c.JSON(http.StatusOK,
					map[string]string{
						"message": "manager",
						"details": content,
					},
				)
			}
		}
	} else {
		c.JSON(http.StatusInternalServerError, utils.HandleError("Unable to get logs", err, c))
	}
}

func getFileContent(path string) (string, error) {
	log.Debug("Read ", path)
	bytes, err := getBinaryFileContent(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func getBinaryFileContent(path string) ([]byte, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []byte{}, nil
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return []byte{}, err
	}
	return bytes, nil
}
