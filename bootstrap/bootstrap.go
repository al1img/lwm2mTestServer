package bootstrap

import (
	"errors"
	"net"
	"strings"

	"github.com/dustin/go-coap"
	log "github.com/sirupsen/logrus"
)

const maxPktLen = 1500

// Instance bootstrap instance
type Instance struct {
	mux  *coap.ServeMux
	addr string

	clients map[string]client
}

type client struct {
	conn *net.UDPConn
	addr *net.UDPAddr
	buf  []byte
}

var errClientNotFound = errors.New("client not found")

// New creates new bootstrap server
func New(addr string) (instance *Instance) {
	instance = &Instance{mux: coap.NewServeMux(), addr: addr, clients: make(map[string]client)}

	log.Debugf("New bootstrap server: %s", instance.addr)

	instance.mux.Handle("/bs", coap.FuncHandler(instance.bootstrapHandler))

	return instance
}

// Start starts bootstram server
func (instance *Instance) Start() {
	log.Debugf("Start bootstrap server: %s", instance.addr)

	go coap.ListenAndServe("udp", instance.addr, instance.mux)
}

// GetClients returns list of clients
func (instance *Instance) GetClients() (clients []string) {
	clients = make([]string, 0, len(instance.clients))

	for client := range instance.clients {
		clients = append(clients, client)
	}

	return clients
}

// Discover bootstrap discover
func (instance *Instance) Discover(name, path string) (result string, err error) {
	log.Debugf("Bootstrap discover, client: %s, path: %s", name, path)

	client, ok := instance.clients[name]
	if !ok {
		return "", errClientNotFound
	}

	request := coap.Message{
		Type: coap.Confirmable,
		Code: coap.GET}

	if path != "/" {
		request.SetPathString(path)
	}

	request.SetOption(coap.Accept, coap.AppLinkFormat)

	response, err := client.send(request)
	if err != nil {
		return "", err
	}

	if response.Code != coap.Content {
		return "", errors.New(response.Code.String())
	}

	return string(response.Payload), nil
}

func (instance *Instance) bootstrapHandler(conn *net.UDPConn, addr *net.UDPAddr, message *coap.Message) *coap.Message {
	var ep string

	queries := message.Options(coap.URIQuery)

	for _, query := range queries {
		if strings.HasPrefix(query.(string), "ep=") {
			ep = strings.TrimPrefix(query.(string), "ep=")
		}
	}

	log.Infof("Bootstrap request ep = %s", ep)

	instance.clients[ep] = client{conn, addr, make([]byte, maxPktLen)}

	response := &coap.Message{
		Type:      coap.Acknowledgement,
		Code:      coap.Changed,
		MessageID: message.MessageID,
		Token:     message.Token}

	return response
}

func (client *client) send(request coap.Message) (response *coap.Message, err error) {
	if err = coap.Transmit(client.conn, client.addr, request); err != nil {
		return nil, err
	}

	if !request.IsConfirmable() {
		return nil, nil
	}

	message, err := coap.Receive(client.conn, client.buf)
	if err != nil {
		return nil, err
	}

	return &message, nil
}
