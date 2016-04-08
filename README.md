go-dota2
========

A Dota 2 client for Go.

## Example
This code snippet is connecting to Steam, and then connecting to the Dota 2 GC, requesting match details for a sample match and printing the name of the player in slot #1.


## Tests
Steam account credentials for go tests are set via environmental variables, so to run them you'd enter the following command, for example: `STEAM_USERNAME="test_account" STEAM_PASSWORD="test_password" go test`
Also, Steam Log on may be denied, and you may then receive an email with SteamGuard auth code.
Enter that code using: `STEAM_USERNAME="test_account" STEAM_PASSWORD="test_password" STEAM_AUTH_CODE="my_auth_code" go test`
In the output you will see the Steam sentry that was generated for you. From then on, for this account on this machine, you may use your sentry when running test like: `STEAM_USERNAME="test_account" STEAM_PASSWORD="test_password" STEAM_SENTRY="my_sentry" go test`

```go
package main

import (
	"log"

	dota2 "github.com/vvekic/go-dota2"
)

func main() {
	dotaClient := dota2.NewClient()
	if err := dotaClient.Connect("my_username", "my_password", "my_sentry", ""); err != nil {
		log.Fatal(err)
	}
	matchDetails, err := dotaClient.MatchDetails(854233753)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("player slot 1: %s", matchDetails.GetMatch().Players[0].GetPlayerName())
}
```
