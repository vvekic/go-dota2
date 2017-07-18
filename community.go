package dota2

import (
	"fmt"

	"github.com/vvekic/go-steam/dota/protocol/protobuf"
	"github.com/vvekic/go-steam/protocol/gamecoordinator"
)

// ProfileCard sends a request to the Dota 2 GC requesting profile card for account with given ID
func (c *Client) ProfileCard(accountID uint32) (*protobuf.CMsgDOTAProfileCard, error) {
	if !c.gcReady {
		return nil, fmt.Errorf("GC not readyss")
	}

	// log.Printf("Requesting profile card for account ID: %d", accountID)
	requestName := true
	msg := gamecoordinator.NewGCMsgProtobuf(AppId, uint32(protobuf.EDOTAGCMsg_k_EMsgClientToGCGetProfileCard), &protobuf.CMsgDOTAProfileRequest{
		AccountId:   &accountID,
		RequestName: &requestName,
		// Engine:
	})

	response := new(protobuf.CMsgDOTAProfileCard)
	packet, err := c.runJob(msg)
	if err != nil {
		return nil, err
	}
	packet.ReadProtoMsg(response) // Interpret GCPacket and populate `response` with data
	return response, nil
}
