package dota2

import (
	"github.com/Philipp15b/go-steam"
	. "github.com/Philipp15b/go-steam/internal/gamecoordinator"
	"github.com/rjacksonm1/go-dota2/internal/protobuf"
	"log"
	"time"
)

const VERSION = "0.0.1"
const AppId = 570

var helloTicker *time.Ticker

// To use any methods of this, you'll need to SetPlaying(true) and wait for
// the GCReadyEvent.
type Dota2 struct {
	client  *steam.Client
	gcReady bool
	Debug   bool
}

// Creates a new Dota2 instance and registers it as a packet handler
func New(client *steam.Client) *Dota2 {
	d2 := &Dota2{
		client:  client,
		gcReady: false,
		Debug:   false,
	}
	client.GC.RegisterPacketHandler(d2)
	return d2
}

// Tells Steam we're playing Dota 2, and sends ClientHello to request a connection to the Dota 2 GC
func (d2 *Dota2) SetPlaying(playing bool) {
	if playing {
		if d2.Debug {
			log.Print("Setting GamesPlayed to Dota 2")
		}
		d2.client.GC.SetGamesPlayed(AppId)

		// Send ClientHello every 5 seconds.  This ticker will be stopped when we get ClientWelcome from the GC
		helloTicker = time.NewTicker(5 * time.Second)
		go func() {
			for t := range helloTicker.C {

				if d2.Debug {
					log.Print("Sending ClientHello, ", t)
				}

				d2.client.GC.Write(NewGCMsgProtobuf(
					AppId,
					uint32(protobuf.EGCBaseClientMsg_k_EMsgGCClientHello),
					&protobuf.CMsgClientHello{}))
			}
		}()
	} else {
		log.Print("Setting GamesPlayed to nil")
		d2.client.GC.SetGamesPlayed()
	}
}

type GCReadyEvent struct{}

// Handles all GC packets that come from Steam.
// Ignores packets unrelated to Dota 2.
// Routes certain packets to their handlers - if we have handlers defined for them
func (d2 *Dota2) HandleGCPacket(packet *GCPacket) {
	if packet.AppId != AppId {
		return
	}
	switch protobuf.EGCBaseClientMsg(packet.MsgType) {
	case protobuf.EGCBaseClientMsg_k_EMsgGCClientWelcome:
		if d2.Debug {
			log.Print("Received ClientWelcome")
		}
		d2.handleWelcome(packet)

	default:
		log.Print("Recieved GC message without a handler, ",
			packet.MsgType)
	}
}

func (d2 *Dota2) handleWelcome(packet *GCPacket) {
	// Stop sending "Hello"
	if d2.Debug {
		log.Print("Stopping ClientHello ticker")
	}
	helloTicker.Stop()

	if d2.Debug {
		log.Print("Emitting GCReadyEvent")
	}
	d2.client.Emit(&GCReadyEvent{})
}
