package lwm2m

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	coap "github.com/go-ocf/go-coap"
	log "github.com/sirupsen/logrus"
)

/*******************************************************************************
 * Types
 ******************************************************************************/

// Instance lwm2m instance
type Instance struct {
	mux      *coap.ServeMux
	addr     string
	location int

	clients map[string]*Client
}

// Client client description
type Client struct {
	conn         *coap.ClientConn
	location     string
	lt           int
	timer        *time.Timer
	closeChannel chan bool
	objects      string
}

/*******************************************************************************
 * Vars
 ******************************************************************************/

var errClientNotFound = errors.New("client not found")

/*******************************************************************************
 * Public
 ******************************************************************************/

// New creates new lwm2m server
func New(addr string) (instance *Instance) {
	instance = &Instance{mux: coap.NewServeMux(), addr: addr, clients: make(map[string]*Client)}

	log.Debugf("New lwm2m server: %s", instance.addr)

	instance.mux.Handle("/rd", coap.HandlerFunc(instance.registrationHandler))

	return instance
}

// Start starts lwm2m server
func (instance *Instance) Start() {
	log.Debugf("Start lwm2m server: %s", instance.addr)

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

// Discover device discover
func (instance *Instance) Discover(name, path string) (result string, err error) {
	log.Debugf("Device discover, client: %s, path: %s", name, path)

	client, ok := instance.clients[name]
	if !ok {
		return "", errClientNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req, err := client.conn.NewGetRequest(path)
	if err != nil {
		return "", err
	}

	req.AddOption(coap.Accept, coap.AppLinkFormat)

	rsp, err := client.conn.ExchangeWithContext(ctx, req)
	if err != nil {
		return "", err
	}

	if rsp.Code() != coap.Content {
		return "", errors.New(rsp.Code().String())
	}

	return string(rsp.Payload()), nil
}

// Read device read
func (instance *Instance) Read(name, path string) (result string, err error) {
	log.Debugf("Device read, client: %s, path: %s", name, path)

	client, ok := instance.clients[name]
	if !ok {
		return "", errClientNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req, err := client.conn.NewGetRequest(path)
	if err != nil {
		return "", err
	}

	req.AddOption(coap.Accept, 110)

	rsp, err := client.conn.ExchangeWithContext(ctx, req)
	if err != nil {
		return "", err
	}

	if rsp.Code() != coap.Content {
		return "", errors.New(rsp.Code().String())
	}

	return string(rsp.Payload()), nil
}

/*******************************************************************************
 * Private
 ******************************************************************************/

func (instance *Instance) registrationHandler(w coap.ResponseWriter, req *coap.Request) {
	if req.Msg.Code() != coap.POST {
		log.Errorf("Wrong request code: %s", req.Msg.Code().String())
		return
	}

	var ep string
	var lt int
	var err error

	queries := req.Msg.Query()

	for _, query := range queries {
		if strings.HasPrefix(query, "ep=") {
			ep = strings.TrimPrefix(query, "ep=")
		}

		if strings.HasPrefix(query, "lt=") {
			lt, err = strconv.Atoi(strings.TrimPrefix(query, "lt="))
			if err != nil {
				log.Errorf("Can't get lifetime: %s", err)
				return
			}
		}
	}

	log.Infof("Registration request ep = %s, lt = %d", ep, lt)

	rsp := w.NewResponse(coap.Created)

	if err := instance.createClient(req.Client, ep, lt); err != nil {
		log.Errorf("Can't create client: %s", err)
		rsp.SetCode(coap.BadRequest)
	} else {
		instance.clients[ep].objects = string(req.Msg.Payload())
		rsp.AddOption(coap.LocationPath, instance.clients[ep].location)
		log.Infof("Objects %s", instance.clients[ep].objects)
	}

	ctx, cancel := context.WithTimeout(req.Ctx, time.Second)
	defer cancel()

	if err := w.WriteMsgWithContext(ctx, rsp); err != nil {
		log.Errorf("Cannot send response: %s", err)
	}
}

func (instance *Instance) registrationUpdate(ep string, w coap.ResponseWriter, req *coap.Request) {
	switch req.Msg.Code() {
	case coap.POST:
		w.SetCode(coap.Changed)

		queries := req.Msg.Query()

		for _, query := range queries {
			if strings.HasPrefix(query, "lt=") {
				lt, err := strconv.Atoi(strings.TrimPrefix(query, "lt="))
				if err != nil {
					log.Errorf("Can't get lifetime: %s", err)
					w.SetCode(coap.InternalServerError)
					break
				}

				instance.clients[ep].lt = lt
			}
		}

		instance.clients[ep].timer.Reset(time.Duration(instance.clients[ep].lt) * time.Second)

		log.Infof("Registration update ep = %s, lt = %d", ep, instance.clients[ep].lt)
		if len(req.Msg.Payload()) > 0 {
			instance.clients[ep].objects = string(req.Msg.Payload())
			log.Infof("Objects %s", instance.clients[ep].objects)
		}

	case coap.DELETE:
		log.Infof("Deregistration ep = %s", ep)

		w.SetCode(coap.Changed)

		instance.clients[ep].closeChannel <- true

		if err := instance.deleteClient(ep); err != nil {
			log.Errorf("Can't delete client: %s", err)
			w.SetCode(coap.InternalServerError)
		}

	default:
		w.SetCode(coap.BadRequest)
	}

	ctx, cancel := context.WithTimeout(req.Ctx, time.Second)
	defer cancel()

	if _, err := w.WriteWithContext(ctx, nil); err != nil {
		log.Errorf("Cannot send response: %s", err)
	}
}

func (instance *Instance) createClient(conn *coap.ClientConn, ep string, lt int) (err error) {
	if client, ok := instance.clients[ep]; ok {
		client.timer.Stop()
		client.closeChannel <- true
	}

	instance.clients[ep] = &Client{
		conn:         conn,
		timer:        time.NewTimer(time.Duration(lt) * time.Second),
		lt:           lt,
		location:     "rd/" + strconv.Itoa(instance.location),
		closeChannel: make(chan bool)}

	instance.location++

	if err = instance.mux.Handle(instance.clients[ep].location, coap.HandlerFunc(func(w coap.ResponseWriter, req *coap.Request) {
		instance.registrationUpdate(ep, w, req)
	})); err != nil {
		return err
	}

	go func() {
		select {
		case <-instance.clients[ep].timer.C:
			log.Errorf("Client %s registration update timeout", ep)

			if err = instance.deleteClient(ep); err != nil {
				log.Errorf("Can't delete client: %s", err)
				return
			}

		case <-instance.clients[ep].closeChannel:
		}
	}()

	return nil
}

func (instance *Instance) deleteClient(ep string) (err error) {
	log.Debugf("Delete client: %s", ep)

	if _, ok := instance.clients[ep]; !ok {
		return errClientNotFound
	}

	instance.clients[ep].timer.Stop()

	err = instance.mux.HandleRemove(instance.clients[ep].location)
	delete(instance.clients, ep)

	return err
}
