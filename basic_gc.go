package dota2

import (
	. "github.com/Philipp15b/go-steam/internal/gamecoordinator"
	"github.com/rjacksonm1/go-dota2/internal/protobuf"
	"log"
	"time"
)

var helloTicker *time.Ticker

// Namespace for very basic GC calls, e.g. initializing a Dota 2 GC connection
type BasicGC struct {
	d2 *Dota2
}

// Continually send "Hello" to the Dota 2 GC to initialize a connection.  Will send hello every 5 seconds until the GC responds with "Welcome"
func (gc *BasicGC) sendHello() {
	// Send ClientHello every 5 seconds.  This ticker will be stopped when we get ClientWelcome from the GC
	helloTicker = time.NewTicker(5 * time.Second)
	go func() {
		for t := range helloTicker.C {

			if gc.d2.Debug {
				log.Print("Sending ClientHello, ", t)
			}

			gc.d2.client.GC.Write(NewGCMsgProtobuf(
				AppId,
				uint32(protobuf.EGCBaseClientMsg_k_EMsgGCClientHello),
				&protobuf.CMsgClientHello{}))
		}
	}()
}

// Handle the GC's "Welcome" message; stops the "Hello" ticker and emits GCReadyEvent.
func (gc *BasicGC) handleWelcome(packet *GCPacket) {
	// Stop sending "Hello"
	if gc.d2.Debug {
		log.Print("Stopping ClientHello ticker")
	}
	helloTicker.Stop()

	if gc.d2.Debug {
		log.Print("Emitting GCReadyEvent")
	}
	gc.d2.gcReady = true
	gc.d2.client.Emit(&GCReadyEvent{})
}
