package dota2

import (
	"encoding/base64"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/vvekic/go-steam"
	"github.com/vvekic/go-steam/dota/protocol/protobuf"
	"github.com/vvekic/go-steam/protocol"
	"github.com/vvekic/go-steam/protocol/gamecoordinator"
	"github.com/vvekic/go-steam/protocol/steamlang"
)

const AppId = 570

var readyTimeout time.Duration = time.Second * 30

// To use any methods of this, you'll need to SetPlaying(true) and wait for
// the GCReadyEvent.
type Client struct {
	sc        *steam.Client // Steam client, of course!
	readyLock sync.Mutex
	readyChan chan struct{}
	gcReady   bool // Used internally to prevent sending GC reqs when we don't have a GC connection
	creds     *steam.LogOnDetails

	jobs      map[protocol.JobId]chan *gamecoordinator.GCPacket // Set of channels.  Used to sync up go-steam's event-based GC calls.
	lastJobID protocol.JobId                                    // Last job ID. We will increment this for each job we create
	jobsLock  sync.RWMutex
}

// Creates a new Dota2 instance and registers it as a packet handler
func NewClient() *Client {
	c := &Client{
		sc:        steam.NewClient(),
		readyChan: make(chan struct{}),
		gcReady:   false,
		jobs:      make(map[protocol.JobId]chan *gamecoordinator.GCPacket),
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
			log.Printf("Received ClientWelcome")
			c.handleWelcome(packet)
		case int32(protobuf.EGCBaseClientMsg_k_EMsgGCClientConnectionStatus):
			c.handleConnectionStatus(packet)
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
	helloTicker = time.NewTicker(5 * time.Second)
	go func() {
		for range helloTicker.C {
			log.Printf("Sending ClientHello")
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

func (c *Client) Connect(username, password, sentry, authCode string) error {
	// Check for credentials
	c.creds = &steam.LogOnDetails{
		Username: username,
		Password: password,
		AuthCode: authCode,
	}
	if sentry != "" {
		decodedSentry, err := base64.StdEncoding.DecodeString(sentry)
		if err != nil {
			log.Printf("error decoding sentry")
		} else {
			c.creds.SentryFileHash = decodedSentry
		}
	}
	if c.creds.Username == "" {
		return fmt.Errorf("username not set")
	}
	if c.creds.Password == "" {
		return fmt.Errorf("password not set")
	}
	c.sc.Connect()

	select {
	case <-c.readyChan:
		return nil
	case <-time.After(readyTimeout):
		return fmt.Errorf("timeout waiting for GC to become ready")
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
			log.Printf("Connected to Steam")
			c.onSteamConnected()
		case *steam.MachineAuthUpdateEvent:
			log.Printf("Received new sentry: %s", base64.StdEncoding.EncodeToString(e.Hash))
		case *steam.LoggedOnEvent:
			log.Printf("Logged on to Steam")
			c.onSteamLogon()
		case *steam.LogOnFailedEvent:
			log.Printf("Log on failed, result: %s", e.Result)
		case *steam.LoggedOffEvent:
			log.Printf("Logged off, result: %s", e.Result)
		case *steam.DisconnectedEvent:
			log.Printf("Disconnected from Steam.")
		case *steam.AccountInfoEvent:
			log.Printf("Account name: %s, Country: %s, Authorized machines: %d", e.PersonaName, e.Country, e.CountAuthedComputers)
		case *steam.LoginKeyEvent:
			log.Printf("Login Key: %s", e.LoginKey)
		case *steam.WebSessionIdEvent, *steam.PersonaStateEvent, *steam.FriendsListEvent, *steam.ClientCMListEvent:
			// mute

			// custom events
		case *GCReadyEvent:
			c.readyLock.Lock()
			c.readyChan <- struct{}{}
			c.readyLock.Unlock()
			log.Printf("Dota 2 Game Coordinator ready!")

			// errors
		case steam.FatalErrorEvent, error:
			log.Fatal(e)

		default:
			log.Printf("unknown steam event: %#v", e)
		}
	}
}

func (c *Client) onSteamConnected() {
	log.Printf("Logging on as %s", c.creds.Username)
	c.sc.Auth.LogOn(c.creds)
}

func (c *Client) onSteamLogon() {
	log.Printf("Setting social status to Offline")
	// Set steam social status to 'offline')
	c.sc.Social.SetPersonaState(steamlang.EPersonaState_Offline)

	// Launch Dota 2
	c.setPlaying(true)
	c.sendHello()
}

func (c *Client) onDotaGCReady() {
	log.Print("Dota 2 GC ready!")
	matchDeets, err := c.MatchDetails(854233753)
	if err != nil {
		log.Printf("error getting match details: %v", err)
		return
	}

	log.Printf("Got match details for match id %d", matchDeets.GetMatch().GetMatchId())
	log.Printf("%v", matchDeets)
}
