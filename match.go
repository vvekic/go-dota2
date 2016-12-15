package dota2

import (
	"fmt"
	"log"

	"github.com/vvekic/go-steam/dota/protocol/protobuf"
	"github.com/vvekic/go-steam/protocol/gamecoordinator"
)

func (c *Client) MatchDetailsPar(ids []int) []*protobuf.CMsgGCMatchDetailsResponse {
	log.Printf("matchdetails %v", ids)
	var responses []*protobuf.CMsgGCMatchDetailsResponse
	out := make(chan *protobuf.CMsgGCMatchDetailsResponse)
	defer close(out)
	for _, id := range ids {
		go func(id int) {
			res, err := c.MatchDetails(uint64(id))
			if err != nil {
				log.Printf("client %d MatchDetails error: %v", c.Id, err)
				out <- nil
				return
			}
			out <- res
		}(id)
	}
	for i := 0; i < len(ids); i++ {
		r := <-out
		if r != nil {
			responses = append(responses, r)
		}
	}
	return responses
}

// Sends a request to the Dota 2 GC requesting details for the given matchid.
func (c *Client) ServerMatchDetails(matchIDs []uint64) (*protobuf.CMsgGCToServerMatchDetailsResponse, error) {
	if !c.gcReady {
		return nil, fmt.Errorf("GC not ready")
	}

	log.Printf("Requesting match details for match IDs: %v", matchIDs)
	msgToGC := gamecoordinator.NewGCMsgProtobuf(
		AppId,
		uint32(protobuf.EDOTAGCMsg_k_EMsgServerToGCMatchDetailsRequest),
		&protobuf.CMsgServerToGCMatchDetailsRequest{
			MatchIds: matchIDs,
		})

	response := new(protobuf.CMsgGCToServerMatchDetailsResponse)
	packet, err := c.runJob(msgToGC)
	if err != nil {
		return nil, err
	}
	packet.ReadProtoMsg(response) // Interpret GCPacket and populate `response` with data
	return response, nil
}

// Sends a request to the Dota 2 GC requesting details for the given matchid.
func (c *Client) MatchDetails(matchID uint64) (*protobuf.CMsgGCMatchDetailsResponse, error) {
	if !c.gcReady {
		return nil, fmt.Errorf("GC not ready")
	}

	log.Printf("Requesting match details for match ID: %d", matchID)

	msgToGC := gamecoordinator.NewGCMsgProtobuf(
		AppId,
		uint32(protobuf.EDOTAGCMsg_k_EMsgGCMatchDetailsRequest),
		&protobuf.CMsgGCMatchDetailsRequest{
			MatchId: &matchID,
		})

	response := new(protobuf.CMsgGCMatchDetailsResponse)
	packet, err := c.runJob(msgToGC)
	if err != nil {
		return nil, err
	}
	packet.ReadProtoMsg(response) // Interpret GCPacket and populate `response` with data
	return response, nil
}

func (c *Client) Matches(startMatchID int, matchesRequested uint32) (*protobuf.CMsgDOTARequestMatchesResponse, error) {
	if !c.gcReady {
		return nil, fmt.Errorf("GC not ready")
	}

	log.Printf("Requesting matches starting at match ID: %d", startMatchID)

	req := &protobuf.CMsgDOTARequestMatches{
		MatchesRequested: &matchesRequested,
	}
	if startMatchID >= 0 {
		var id uint64
		id = uint64(startMatchID)
		req.StartAtMatchId = &id
	}

	msgToGC := gamecoordinator.NewGCMsgProtobuf(
		AppId,
		uint32(protobuf.EDOTAGCMsg_k_EMsgGCRequestMatches),
		req,
	)

	response := new(protobuf.CMsgDOTARequestMatchesResponse)
	packet, err := c.runJob(msgToGC)
	if err != nil {
		return nil, err
	}
	packet.ReadProtoMsg(response) // Interpret GCPacket and populate `response` with data
	return response, nil
}

func (c *Client) MatchMinimalDetails(matchIDs ...uint64) (*protobuf.CMsgClientToGCMatchesMinimalResponse, error) {
	if !c.gcReady {
		return nil, fmt.Errorf("GC not ready")
	}

	log.Printf("Requesting minimal match details for match IDs: %v", matchIDs)

	msgToGC := gamecoordinator.NewGCMsgProtobuf(
		AppId,
		uint32(protobuf.EDOTAGCMsg_k_EMsgClientToGCMatchesMinimalRequest),
		&protobuf.CMsgClientToGCMatchesMinimalRequest{
			MatchIds: matchIDs,
		})

	response := new(protobuf.CMsgClientToGCMatchesMinimalResponse)
	packet, err := c.runJob(msgToGC)
	if err != nil {
		return nil, err
	}
	packet.ReadProtoMsg(response) // Interpret GCPacket and populate `response` with data
	return response, nil
}
