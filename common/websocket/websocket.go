package websocket

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/soprasteria/intools-engine/common/utils"
)

const (
	defaultChannelLength = 100
)

var (
	appclient = &AppClient{
		Clients: make(map[*websocket.Conn]*Client),
	}
	wsupgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	ConnectorBuffer chan *LightConnector
)

type LightConnector struct {
	GroupId     string
	ConnectorId string
	Value       *map[string]interface{}
}

type Client struct {
	Socket   *websocket.Conn
	GroupIds []string
}

type AppClient struct {
	Clients map[*websocket.Conn]*Client
}

type Message struct {
	Key  string                 `json:"key"`
	Data map[string]interface{} `json:"data"`
}

func InitChannel(length int64) {
	if length <= 0 {
		length = defaultChannelLength
	}
	ConnectorBuffer = make(chan *LightConnector, length)
	log.Info("Initializing websocket buffered channel with a size of ", length)
	go func() {
		for {
			notify(<-ConnectorBuffer)
		}
	}()
}

// Get websocket
func GetWS(c *gin.Context) {
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.WithError(err).Error("Failed to set websocket upgrade")
		return
	}
	// Registering the connection from Intools back-office
	err = appclient.Register(conn)
	if err != nil {
		switch err.(type) {
		case *websocket.CloseError:
			log.Info("Communication with client has been interrupted : websocket closed")
			return
		default:
			log.WithError(err).Error("Error while registering client, closing websocket")
			conn.Close()
		}
	}
}

// Registers the connection between a client (one intools backoffice instance) and server (this intools engine instance)
func (appClient *AppClient) Register(conn *websocket.Conn) error {

	client := createClient(conn)
	appclient.bindClient(conn, &client)
	err := sendAck(conn)
	if err != nil {
		log.WithError(err).Error("Can't send ack to the client")
		return err
	}

	log.WithField("client", client).Info("Client is now registered to engine")

	err = appclient.handleEvents(conn, &client)
	if err != nil {
		return err
	}
	return nil
}

// Broadcasts the value to all client registered to the group
func notify(lConnector *LightConnector) {
	log.WithField("groupID", lConnector.GroupId).Info("Notifying all client registered")
	log.WithFields(log.Fields{"value": lConnector.Value, "clients": appclient.Clients}).Debug("Send value to clients")

	for _, client := range appclient.Clients {
		log.WithField("groupIDs", client.GroupIds).Debug("Clients")
		if utils.Contains(client.GroupIds, lConnector.GroupId) {
			message := createConnectorValueMessage(lConnector.ConnectorId, lConnector.Value)
			log.WithFields(log.Fields{"client": client, "message": message}).Debug("Notifying client with message")
			err := client.Socket.WriteJSON(message)
			if err != nil {
				log.WithError(err).Warnf("Can't send connector value of id %s, to client %p", lConnector.ConnectorId, client)
			}
		}
	}
}

// Creates message, structured as value send to the client
func createConnectorValueMessage(connectorId string, value *map[string]interface{}) Message {
	data := map[string]interface{}{
		"connectorId": connectorId,
		"value":       value,
	}
	message := Message{
		Key:  "connector-value",
		Data: data,
	}
	return message
}

// Create a simple client
func createClient(conn *websocket.Conn) Client {
	log.Debug("Connection event from client")
	var client = &Client{
		Socket:   conn,
		GroupIds: []string{},
	}
	return *client
}

// Add the client to the connected clients
func (appClient *AppClient) bindClient(conn *websocket.Conn, c *Client) {
	log.Debugf("clients before %v", appClient.Clients)
	appClient.Clients[conn] = c
	log.Debugf("clients %v", appClient.Clients)
}

// Send ack to the client
func sendAck(conn *websocket.Conn) error {
	message := Message{
		Key:  "connected",
		Data: nil,
	}
	err := conn.WriteJSON(message)
	if err != nil {
		return err
	}
	return nil
}

// Handle events from client
func (appClient *AppClient) handleEvents(conn *websocket.Conn, client *Client) error {
Events:
	for {
		// Read message
		var message Message
		err := conn.ReadJSON(&message)
		if err != nil {
			switch err.(type) {
			case *websocket.CloseError:
				log.Debugf("Websocket %p is deconnected. Removing from clients", conn)
				delete(appclient.Clients, conn)
				log.Debugf("Clients are now  %v", appClient.Clients)
				return err
			default:
				log.WithError(err).Warn("Error while reading json message")
				continue Events
			}
		}

		log.Debugf("Message %v", message)

		// Check message structure
		id, ok := message.Data["groupId"]
		if !ok {
			log.Warn("Can't register or unregister group because groupId does not exist in message")
			continue Events
		}
		groupId, ok := id.(string)
		if !ok {
			log.Warnf("Can't register or unregister group because groupId is not string : %s", groupId)
			continue Events
		}

		// Handles types of messages
		switch message.Key {
		case "register-group":
			// Handles group registering for the client

			client.GroupIds = append(client.GroupIds, groupId)
			log.WithField("group", groupId).WithField("client", client).Info("Group registered to engine")
		case "unregister-group":
			// Handles group unregistering for the client
			i, ok := utils.IndexOf(client.GroupIds, groupId)
			if ok {
				client.GroupIds = append(client.GroupIds[:i], client.GroupIds[i+1:]...)
				log.WithField("group", groupId).WithField("client", client).Info("Group is unregistered from engine")
			}
		}
		log.Debugf("Registered groups for client %p are now : %s", client, client.GroupIds)
	}
}
