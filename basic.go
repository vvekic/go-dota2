package dota2

import (
	"fmt"
	"log"
	"time"

	"github.com/Philipp15b/go-steam/dota/protocol/protobuf"
	"github.com/Philipp15b/go-steam/protocol"
	"github.com/Philipp15b/go-steam/protocol/gamecoordinator"
)

var (
	helloTicker *time.Ticker
	jobTimeout  time.Duration = time.Second * 10
)

func (c *Client) runJob(msgToGC *gamecoordinator.GCMsgProtobuf) (*gamecoordinator.GCPacket, error) {
	// Create job ID for this request
	jobId := protocol.JobId(c.lastJobID + 1)
	c.lastJobID = jobId
	msgToGC.SetSourceJobId(jobId)

	// Create a channel for this job
	c.jobsLock.Lock()
	jobChan := make(chan *gamecoordinator.GCPacket)
	c.jobs[jobId] = jobChan
	c.jobsLock.Unlock()

	// Write this request to the GC
	c.sc.GC.Write(msgToGC)

	select {
	case packet := <-jobChan: // GCPacket response from GC
		return packet, nil
	case <-time.After(jobTimeout):
		c.jobsLock.Lock()
		delete(c.jobs, jobId)
		close(jobChan)
		c.jobsLock.Unlock()
		return nil, fmt.Errorf("job %d timeout", jobId)
	}
}

type GCReadyEvent struct{}

// Handle the GC's "Welcome" message; stops the "Hello" ticker and emits GCReadyEvent.
func (c *Client) handleWelcome(packet *gamecoordinator.GCPacket) {
	helloTicker.Stop()
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
	log.Printf("Received CacheSubscribed version %v", response.GetVersion())
}
