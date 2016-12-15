package dota2

import (
	"fmt"
	"log"
	"time"

	"github.com/vvekic/go-steam/dota/protocol/protobuf"
	"github.com/vvekic/go-steam/protocol"
	"github.com/vvekic/go-steam/protocol/gamecoordinator"
)

var (
	jobTimeout = time.Second * 10
	jobRetries = 10
)

type timeoutError struct {
	err error
}

func (t timeoutError) IsTimeout() {}

func (t timeoutError) Error() string {
	return t.err.Error()
}

func (c *Client) runJob(msg *gamecoordinator.GCMsgProtobuf) (*gamecoordinator.GCPacket, error) {
	// Create a channel for this job
	jobChan := make(chan *gamecoordinator.GCPacket)

	// Create job ID for this request
	c.jobsLock.Lock()
	// Create job ID for this request
	jobID := protocol.JobId(c.lastJobID + 1)
	c.lastJobID = jobID
	c.jobs[jobID] = jobChan
	c.jobsLock.Unlock()

	msg.SetSourceJobId(jobID)

	// log.Printf("client %d job %d", c.Id, jobId)
	// Write this request to the GC
	c.sc.GC.Write(msg)

	select {
	case packet := <-jobChan: // GCPacket response from GC
		c.jobsLock.Lock()
		delete(c.jobs, jobID)
		c.jobsLock.Unlock()
		return packet, nil
	case <-time.After(jobTimeout):
		c.jobsLock.Lock()
		delete(c.jobs, jobID)
		close(jobChan)
		c.jobsLock.Unlock()
		return nil, timeoutError{fmt.Errorf("job %d timeout", jobID)}
	}
}

type GCReadyEvent struct{}

type DisconnectedEvent struct{}

// Handle the GC's "Welcome" message; stops the "Hello" ticker and emits GCReadyEvent.
func (c *Client) handleWelcome(packet *gamecoordinator.GCPacket) {
	c.helloTicker.Stop()
	c.gcReady = true
	c.sc.Emit(new(GCReadyEvent))
}

func (c *Client) handleConnectionStatus(packet *gamecoordinator.GCPacket) {
	// Construct and wait for the GC's response (will be piped to our jobs channel)
	response := new(protobuf.CMsgConnectionStatus)
	packet.ReadProtoMsg(response) // Interpret GCPacket and populate `response` with data
	log.Printf("Received ConnectionStatus %s", response.GetStatus())
}

func (c *Client) handleCacheSubscribed(packet *gamecoordinator.GCPacket) {
	response := new(protobuf.CMsgSOCacheSubscribed)
	packet.ReadProtoMsg(response) // Interpret GCPacket and populate `response` with data
	log.Printf("Client %d Received CacheSubscribed version %v", c.Id, response.GetVersion())
}

func (c *Client) handleGetEventPointsResponse(packet *gamecoordinator.GCPacket) {
	response := new(protobuf.CMsgDOTAGetEventPointsResponse)
	packet.ReadProtoMsg(response)
	log.Printf("Received GetEventPointsResponse %s", response.String())
}
