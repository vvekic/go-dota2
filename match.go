package dota2

import (
	. "github.com/Philipp15b/go-steam/internal"
	. "github.com/Philipp15b/go-steam/internal/gamecoordinator"
	"github.com/rjacksonm1/go-dota2/internal/protobuf"
	"log"
)

// Serves as a namespace for match-related methods.
type Match struct {
	d2 *Dota2
}

// Sends a request to the Dota 2 GC requesting details for the given matchid.
func (match *Match) RequestDetails(matchid uint32) *protobuf.CMsgGCMatchDetailsResponse {
	if !match.d2.gcReady {
		log.Printf("Cannot request match details for %s.  GC not ready", matchid)
		panic("GC not ready")
	}

	if match.d2.Debug {
		log.Printf("Requesting match details for matchid %d\n", matchid)
	}

	msgToGC := NewGCMsgProtobuf(
		AppId,
		uint32(protobuf.EDOTAGCMsg_k_EMsgGCMatchDetailsRequest),
		&protobuf.CMsgGCMatchDetailsRequest{
			MatchId: &matchid,
		})

	// Create job ID for this request (TODO: Make a wrapper than does this for us?)
	jobId := JobId(match.d2.lastJobID + 1)
	match.d2.lastJobID = jobId
	msgToGC.SetSourceJobId(jobId)

	// Create a channel for this job
	match.d2.jobs[jobId] = make(chan *GCPacket)

	// Write this request to the GC
	match.d2.client.GC.Write(msgToGC)

	// Construct and wait for the GC's response (will be piped to our jobs channel)
	response := new(protobuf.CMsgGCMatchDetailsResponse)
	packet := <-match.d2.jobs[jobId] // GCPacket response from GC
	packet.ReadProtoMsg(response)    // Interpret GCPacket and populate `response` with data

	// TODO: Handle timeouts (GC doesn't respond)

	return response
}
