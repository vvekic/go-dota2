package dota2

import (
	"encoding/base64"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/vvekic/go-steam"
	"github.com/vvekic/go-steam/dota/protocol/protobuf"
	"github.com/vvekic/go-steam/protocol"
	"github.com/vvekic/go-steam/protocol/gamecoordinator"
	"github.com/vvekic/go-steam/protocol/steamlang"
)

const AppId = 570

var readyTimeout time.Duration = time.Second * 30

func init() {
	if err := steam.InitializeSteamDirectory(91); err != nil {
		log.Printf("error initializing Steam Directory, using built-in server list")
	}
}

// To use any methods of this, you'll need to SetPlaying(true) and wait for
// the GCReadyEvent.
type Client struct {
	Id          int
	helloTicker *time.Ticker
	sc          *steam.Client // Steam client, of course!
	readyLock   sync.Mutex
	readyChan   chan struct{}
	quitChan    chan struct{}
	gcReady     bool // Used internally to prevent sending GC reqs when we don't have a GC connection
	Creds       *steam.LogOnDetails
	connected   bool
	Timeouts    int

	jobs      map[protocol.JobId]chan *gamecoordinator.GCPacket // Set of channels.  Used to sync up go-steam's event-based GC calls.
	lastJobID protocol.JobId                                    // Last job ID. We will increment this for each job we create
	jobsLock  sync.RWMutex
}

// Creates a new Dota2 instance and registers it as a packet handler
func NewClient() *Client {
	c := &Client{
		sc:        steam.NewClient(),
		readyChan: make(chan struct{}),
		quitChan:  make(chan struct{}),
		gcReady:   false,
		jobs:      make(map[protocol.JobId]chan *gamecoordinator.GCPacket),
		connected: false,
	}
	c.sc.GC.RegisterPacketHandler(c)
	go c.loop()
	return c
}

// Handles all GC packets that come from Steam and routes them to their relevant handlers.
func (c *Client) HandleGCPacket(packet *gamecoordinator.GCPacket) {
	if packet.AppId != AppId {
		return
	}

	// If we have a handler channel for this, pipe the GC packet straight there,
	// otherwise use our own routing.
	c.jobsLock.RLock()
	handlerChan, ok := c.jobs[packet.TargetJobId]
	c.jobsLock.RUnlock()
	if ok {
		handlerChan <- packet
	} else {
		// All key types are derived from int32, so cast to int32 to allow us to use a single switch for all types.
		switch int32(packet.MsgType) {
		case int32(protobuf.EGCBaseClientMsg_k_EMsgGCClientWelcome):
			log.Printf("Client %d Received ClientWelcome", c.Id)
			c.handleWelcome(packet)
		case int32(protobuf.EGCBaseClientMsg_k_EMsgGCClientConnectionStatus):
			c.handleConnectionStatus(packet)
		case int32(protobuf.EDOTAGCMsg_k_EMsgDOTAGetEventPointsResponse):
			c.handleGetEventPointsResponse(packet)
		case int32(protobuf.ESOMsg_k_ESOMsg_CacheSubscribed):
			c.handleCacheSubscribed(packet)
		default:
			log.Printf("Recieved GC message without a handler, type %d", packet.MsgType)
		}
	}
}

// SetPlaying tells Steam we're playing Dota 2
func (c *Client) setPlaying(playing bool) {
	if playing {
		log.Printf("Setting GamesPlayed to Dota 2")
		c.sc.GC.SetGamesPlayed(AppId)
	} else {
		log.Print("Setting GamesPlayed to nil")
		c.sc.GC.SetGamesPlayed()
	}
}

// sendHello sends ClientHello to request a connection to the Dota 2 GC
// Continually send "Hello" to the Dota 2 GC to initialize a connection.  Will send hello every 5 seconds until the GC responds with "Welcome"
func (c *Client) sendHello() {
	// Send ClientHello every 5 seconds.  This ticker will be stopped when we get ClientWelcome from the GC
	c.helloTicker = time.NewTicker(5 * time.Second)
	go func() {
		for range c.helloTicker.C {
			log.Printf("Client %d Sending ClientHello", c.Id)
			c.sc.GC.Write(gamecoordinator.NewGCMsgProtobuf(
				AppId,
				uint32(protobuf.EGCBaseClientMsg_k_EMsgGCClientHello),
				&protobuf.CMsgClientHello{
					// set engine to Source 2
					Engine: protobuf.ESourceEngine_k_ESE_Source2.Enum(),
				}))
		}
	}()
}

func (c *Client) ConnectWithCreds(creds *steam.LogOnDetails) error {
	if err := c.sc.Connect(); err != nil {
		return errors.Wrap(err, "error connecting to Steam server")
	}

	select {
	case <-c.readyChan:
		return nil
	case <-time.After(readyTimeout):
		return fmt.Errorf("timeout waiting for GC to become ready")
	case <-c.quitChan:
		return fmt.Errorf("client disconnected")
	}
}

func (c *Client) Connect(username, password, sentry, authCode string) error {
	log.Printf("client %d connecting(%s, %s, %s, %s)", c.Id, username, password, sentry, authCode)
	// Check for credentials
	c.Creds = &steam.LogOnDetails{
		Username: username,
		Password: password,
		AuthCode: authCode,
	}
	if sentry != "" {
		decodedSentry, err := base64.StdEncoding.DecodeString(sentry)
		if err != nil {
			log.Printf("error decoding sentry")
		} else {
			c.Creds.SentryFileHash = decodedSentry
		}
	}
	if c.Creds.Username == "" {
		return fmt.Errorf("username not set")
	}
	if c.Creds.Password == "" {
		return fmt.Errorf("password not set")
	}

	return c.ConnectWithCreds(c.Creds)
}

func (c *Client) reconnect() {
	if err := c.ConnectWithCreds(c.Creds); err != nil {
		c.sc.Emit(&steam.DisconnectedEvent{})
	}
}

func (c *Client) Disconnect() {
	c.sc.Disconnect()
}

func (c *Client) loop() {
	for event := range c.sc.Events() {
		switch e := event.(type) {
		// Steam events
		case *steam.ConnectedEvent:
			log.Printf("Client %d Connected to Steam", c.Id)
			c.onSteamConnected()
		case *steam.MachineAuthUpdateEvent:
			log.Printf("Client %d Received new sentry: %s", c.Id, base64.StdEncoding.EncodeToString(e.Hash))
		case *steam.LoggedOnEvent:
			log.Printf("Client %d Logged on to Steam", c.Id)
			c.onSteamLogon()
		case *steam.LogOnFailedEvent:
			log.Printf("Client %d Log on failed, result: %s", c.Id, e.Result)
		case *steam.LoggedOffEvent:
			log.Printf("Client %d Logged off, result: %s", c.Id, e.Result)
		case *steam.DisconnectedEvent:
			log.Printf("Client %d Disconnected from Steam.", c.Id)
			c.gcReady = false
			c.reconnect()
		case *steam.AccountInfoEvent:
			log.Printf("Client %d Account name: %s, Country: %s, Authorized machines: %d, Flags: %s", c.Id, e.PersonaName, e.Country, e.CountAuthedComputers, e.AccountFlags)
		case *steam.LoginKeyEvent:
			log.Printf("Client %d Login Key: %s", c.Id, e.LoginKey)
		case *steam.WebSessionIdEvent, *steam.PersonaStateEvent, *steam.FriendsListEvent:
			// mute
		case *steam.ClientCMListEvent:
			log.Printf("Client %d received CM list (%d servers)", c.Id, len(e.Addresses))
			if e.Addresses != nil && len(e.Addresses) > 0 {
				steam.UpdateSteamDirectory(e.Addresses)
			}

			// custom events
		case *GCReadyEvent:
			c.readyChan <- struct{}{}
			log.Printf("Client %d Dota 2 Game Coordinator ready!", c.Id)

			// errors
		case steam.FatalErrorEvent, error:
			log.Printf("Client %d error: %v", c.Id, e)

		default:
			log.Printf("unknown steam event: %#v", e)
		}
	}
}

func (c *Client) onSteamConnected() {
	log.Printf("Client %d Logging on as %s", c.Id, c.Creds.Username)
	c.sc.Auth.LogOn(c.Creds)
}

func (c *Client) onSteamLogon() {
	log.Printf("Client %d Setting social status to Offline", c.Id)
	// Set steam social status to 'offline')
	c.sc.Social.SetPersonaState(steamlang.EPersonaState_Offline)

	// Launch Dota 2
	c.setPlaying(true)
	c.sendHello()
}
