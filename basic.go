package dota2

import (
	"fmt"
	"log"
	"time"

	"github.com/cenk/backoff"
	"github.com/vvekic/go-steam/dota/protocol/protobuf"
	"github.com/vvekic/go-steam/protocol"
	"github.com/vvekic/go-steam/protocol/gamecoordinator"
)

var (
	helloTicker *time.Ticker
	jobTimeout  time.Duration = time.Second * 6
	jobRetries                = 10
)

type timeoutError struct {
	err error
}

func (t timeoutError) Error() string {
	return t.err.Error()
}

func (c *Client) runJob(msg *gamecoordinator.GCMsgProtobuf) (*gamecoordinator.GCPacket, error) {

	var packet *gamecoordinator.GCPacket

	operation := func() error {
		p, err := c.runJobOne(msg)
		if err != nil {
			log.Printf("error: %v, backing off", err)
			return err
		}
		packet = p
		return nil
	}

	if err := backoff.Retry(operation, backoff.NewExponentialBackOff()); err != nil {
		return nil, fmt.Errorf("error: %v, stopping back-off", err)
	}

	return packet, nil
}

func (c *Client) runJobOne(msg *gamecoordinator.GCMsgProtobuf) (*gamecoordinator.GCPacket, error) {
	// Create a channel for this job
	jobChan := make(chan *gamecoordinator.GCPacket)

	// Create job ID for this request
	c.jobsLock.Lock()
	// Create job ID for this request
	jobId := protocol.JobId(c.lastJobID + 1)
	c.lastJobID = jobId
	c.jobs[jobId] = jobChan
	c.jobsLock.Unlock()

	msg.SetSourceJobId(jobId)

	// Write this request to the GC
	c.sc.GC.Write(msg)

	select {
	case packet := <-jobChan: // GCPacket response from GC
		c.jobsLock.Lock()
		delete(c.jobs, jobId)
		c.jobsLock.Unlock()
		return packet, nil
	case <-time.After(jobTimeout):
		c.jobsLock.Lock()
		delete(c.jobs, jobId)
		close(jobChan)
		c.jobsLock.Unlock()
		return nil, timeoutError{fmt.Errorf("job %d timeout", jobId)}
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
