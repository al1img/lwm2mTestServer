package bootstrap

import (
	"net"
	"strings"

	"github.com/dustin/go-coap"
	log "github.com/sirupsen/logrus"
)

// Instance bootstrap instance
type Instance struct {
	mux  *coap.ServeMux
	addr string

	clients map[string]*net.UDPAddr
}

// New creates new bootstrap server
func New(addr string) (instance *Instance) {
	instance = &Instance{mux: coap.NewServeMux(), addr: addr, clients: make(map[string]*net.UDPAddr)}

	log.Infof("New bootstrap server: %s", instance.addr)

	instance.mux.Handle("/bs", coap.FuncHandler(instance.bootstrapHandler))

	return instance
}

// Start starts bootstram server
func (instance *Instance) Start() {
	log.Infof("Start bootstrap server: %s", instance.addr)

	go coap.ListenAndServe("udp", instance.addr, instance.mux)
}

func (instance *Instance) bootstrapHandler(connection *net.UDPConn, addr *net.UDPAddr, message *coap.Message) *coap.Message {
	var ep string

	queries := message.Options(coap.URIQuery)

	for _, query := range queries {
		if strings.HasPrefix(query.(string), "ep=") {
			ep = strings.TrimPrefix(query.(string), "ep=")
		}
	}

	log.Infof("Bootstrap request ep = %s", ep)

	instance.clients[ep] = addr

	response := &coap.Message{
		Type:      coap.Acknowledgement,
		Code:      coap.Changed,
		MessageID: message.MessageID,
		Token:     message.Token}

	return response
}
