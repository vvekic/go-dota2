go-dota2
========

A go-steam plugin for Dota 2, consider it in alpha state.
Functionality will be ported over from [node-dota2](https://github.com/RJacksonm1/node-dota2) over the coming weeks.

## WARNING
This code is still very much in flux.  Consider the API unstable, and do not use it for any production apps at present.

## Initializing
You must initialize go-dota2 with `dota2.New(steamClient)`, where `steamClient` is an instance of a go-steam client.

You can also enable go-dota2's logging by setting `Debug` to true, e.g.
```go

	dotaClient := dota2.New(steamClient)
	dotaClient.Debug = true
```

## Example
This code snippet is connecting to Steam, and then connecting to the Dota 2 GC.

Steam account credentials in this example are set via environmental variables, so to run it you'd enter the following command, for example: ` STEAM_USERNAME="test_account" STEAM_PASSWORD="test_password" go run example.go`

```go

package main

import (
	"encoding/base64"
	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/internal/steamlang"
	"github.com/rjacksonm1/go-dota2"
	"log"
	"os"
)

var dotaClient *dota2.Dota2
var steamClient *steam.Client

var MATCHID uint32 = 854233753

func onSteamLogon() {
	// Create Dota2 instance
	dotaClient = dota2.New(steamClient)
	dotaClient.Debug = true

	// Set steam social status to 'busy'
	log.Print("Setting Steam persona state to Busy")
	steamClient.Social.SetPersonaState(steamlang.EPersonaState_Busy)

	// Launch Dota 2
	log.Print("Launching Dota 2")
	dotaClient.SetPlaying(true)
}

func onDotaGCReady() {
	log.Print("Doto GC ready!")

	matchDeets := dotaClient.Match.RequestDetails(MATCHID)

	log.Printf("Got match details for match id %d", matchDeets.GetMatch().GetMatchId())
}

func main() {
	// Check for credentials

	steamCredentials := steam.LogOnDetails{}
	steamCredentials.Username = os.Getenv("STEAM_USERNAME")
	steamCredentials.Password = os.Getenv("STEAM_PASSWORD")

	if os.Getenv("STEAM_SENTRY") != "" {
		decoded_sentry, err := base64.StdEncoding.DecodeString(os.Getenv("STEAM_SENTRY"))
		if err == nil {
			steamCredentials.SentryFileHash = decoded_sentry
		}
	}
	if os.Getenv("STEAM_AUTH_CODE") != "" {
		steamCredentials.AuthCode = os.Getenv("STEAM_AUTH_CODE")
	}

	if steamCredentials.Username == "" || steamCredentials.Password == "" {
		panic("Username or Password not set!")
	}

	steamClient = steam.NewClient()
	log.Print("Connecting to Steam")
	steamClient.Connect()

	for event := range steamClient.Events() {
		switch e := event.(type) {

		// Steam events
		case *steam.ConnectedEvent:
			log.Print("Connected to Steam. Logging on")
			steamClient.Auth.LogOn(steamCredentials)

		case *steam.MachineAuthUpdateEvent:
			log.Printf("Received new sentry; logging it: \"%s\"", base64.StdEncoding.EncodeToString(e.Hash))

		case *steam.LoggedOnEvent:
			log.Print("Logged on to Steam")
			onSteamLogon()

		// Dota2 events
		case *dota2.GCReadyEvent:
			log.Print("Received Dota 2 GC Ready event")
			onDotaGCReady()

		// Errors
		case steam.FatalErrorEvent:
		case error:
			log.Print(e)
		}
	}

}


```