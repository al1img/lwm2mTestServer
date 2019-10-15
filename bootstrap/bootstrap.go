package bootstrap

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/go-ocf/go-coap"
	log "github.com/sirupsen/logrus"
)

// Instance bootstrap instance
type Instance struct {
	mux  *coap.ServeMux
	addr string

	clients map[string]*coap.ClientConn
}

var errClientNotFound = errors.New("client not found")

// New creates new bootstrap server
func New(addr string) (instance *Instance) {
	instance = &Instance{mux: coap.NewServeMux(), addr: addr, clients: make(map[string]*coap.ClientConn)}

	log.Debugf("New bootstrap server: %s", instance.addr)

	instance.mux.Handle("/bs", coap.HandlerFunc(instance.bootstrapHandler))

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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req, err := client.NewGetRequest(path)
	if err != nil {
		return "", err
	}

	req.AddOption(coap.Accept, coap.AppLinkFormat)

	rsp, err := client.ExchangeWithContext(ctx, req)
	if err != nil {
		return "", err
	}

	if rsp.Code() != coap.Content {
		return "", errors.New(rsp.Code().String())
	}

	return "", nil
}

func (instance *Instance) bootstrapHandler(w coap.ResponseWriter, req *coap.Request) {
	if req.Msg.Code() != coap.POST {
		log.Errorf("Wrong request code: %s", req.Msg.Code().String())
		return
	}

	var ep string

	queries := req.Msg.Query()

	for _, query := range queries {
		if strings.HasPrefix(query, "ep=") {
			ep = strings.TrimPrefix(query, "ep=")
		}
	}

	log.Infof("Bootstrap request ep = %s", ep)

	ctx, cancel := context.WithTimeout(req.Ctx, time.Second)
	defer cancel()

	w.SetCode(coap.Changed)

	if _, err := w.WriteWithContext(ctx, nil); err != nil {
		log.Errorf("Cannot send response: %s", err)
	}

	instance.clients[ep] = req.Client

	/*
		rsp := w.NewResponse(coap.Changed)
		rsp.SetType(coap.Acknowledgement)
		rsp.SetMessageID(req.Msg.MessageID())
		rsp.SetToken(req.Msg.Token())

		if err := w.WriteMsgWithContext(ctx, rsp); err != nil {
			log.Errorf("Cannot send response: %s", err)
		}
	*/
}

func (instance *Instance) receiveHandler(conn *net.UDPConn, addr *net.UDPAddr, message *coap.Message) *coap.Message {
	log.Info("Message received")

	return nil
}
