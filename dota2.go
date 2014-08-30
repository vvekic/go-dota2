package dota2

import (
	"github.com/Philipp15b/go-steam"
	. "github.com/Philipp15b/go-steam/internal/gamecoordinator"
	"github.com/rjacksonm1/go-dota2/internal/protobuf"
	"log"
)

const VERSION = "0.0.2"
const AppId = 570

// To use any methods of this, you'll need to SetPlaying(true) and wait for
// the GCReadyEvent.
type Dota2 struct {
	client  *steam.Client
	gcReady bool // Used internally to prevent sending GC reqs when we don't have a GC connection
	Debug   bool // Enabled additional logging

	BasicGC *BasicGC
}

// Creates a new Dota2 instance and registers it as a packet handler
func New(client *steam.Client) *Dota2 {
	d2 := &Dota2{
		client:  client,
		gcReady: false,
		Debug:   false,
	}
	client.GC.RegisterPacketHandler(d2)

	d2.BasicGC = &BasicGC{d2: d2}
	d2.client.GC.RegisterPacketHandler(d2.BasicGC)

	return d2
}

// Tells Steam we're playing Dota 2, and sends ClientHello to request a connection to the Dota 2 GC
func (d2 *Dota2) SetPlaying(playing bool) {
	if playing {
		if d2.Debug {
			log.Print("Setting GamesPlayed to Dota 2")
		}
		d2.client.GC.SetGamesPlayed(AppId)

		// Send hello to GC to initialize GC connection
		d2.BasicGC.sendHello()
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

	default:
		log.Print("Recieved GC message without a handler, ",
			packet.MsgType)
	}
}
