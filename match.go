package dota2

import (
	. "github.com/Philipp15b/go-steam/internal/gamecoordinator"
	"github.com/rjacksonm1/go-dota2/internal/protobuf"
	"log"
)

// Serves as a namespace for match-related methods.
type Match struct {
	d2 *Dota2
}

type MatchDetailsResponseEvent struct {
	body *protobuf.CMsgGCMatchDetailsResponse
}

// Sends a request to the Dota 2 GC requesting details for the given matchid.
// TODO: Note return methods (straight return vs events?)
func (match *Match) RequestDetails(matchid uint32) {
	if !match.d2.gcReady {
		log.Printf("Cannot request match details for %s.  GC not ready", matchid)
		return
	}

	if match.d2.Debug {
		log.Printf("Requesting match details for matchid %s\n", matchid)
	}

	msgToGC := NewGCMsgProtobuf(
		AppId,
		uint32(protobuf.EDOTAGCMsg_k_EMsgGCMatchDetailsRequest),
		&protobuf.CMsgGCMatchDetailsRequest{
			MatchId: &matchid,
		})

	msgToGC.SetSourceJobId(123456) // TODO: Make a GC.Write wrapper that gives each write a unique job ID.

	match.d2.client.GC.Write(msgToGC)
}

// Interprets the GC's response to a match details request, and returns / throws an event as necessary.
func (match *Match) handleDetailsResponse(packet *GCPacket) {
	if match.d2.Debug {
		log.Print("Emitting MatchDetailsResponseEvent.")
	}

	// FIXME: This seems redundant? Is there a simpler way to return this as a clearly identifiable event
	response := MatchDetailsResponseEvent{
		body: new(protobuf.CMsgGCMatchDetailsResponse),
	}
	packet.ReadProtoMsg(response.body)

	log.Printf("Response target job ID: %s", packet.TargetJobId) // TODO: Write an handler that deals with job shit.

	match.d2.client.Emit(&response)
}
