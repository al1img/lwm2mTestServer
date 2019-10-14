package main

import (
	"log"

	"github.com/al1img/lwm2mTestServer/registration"
	coap "github.com/dustin/go-coap"
)

func main() {
	mux := coap.NewServeMux()
	mux.Handle("/rd", coap.FuncHandler(registration.Handler))

	log.Fatal(coap.ListenAndServe("udp", ":5683", mux))
}
