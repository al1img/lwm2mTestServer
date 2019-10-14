package registration

import (
	"net"

	coap "github.com/dustin/go-coap"
	log "github.com/sirupsen/logrus"
)

// Handler registration handler
func Handler(connection *net.UDPConn, addr *net.UDPAddr, message *coap.Message) *coap.Message {
	log.Info("Registration received")

	response := &coap.Message{
		Type:      coap.Acknowledgement,
		Code:      coap.Created,
		MessageID: message.MessageID,
		Token:     message.Token}

	response.SetOption(coap.LocationPath, "/rd/1")

	return response
}
