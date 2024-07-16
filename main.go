package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/fiatjaf/eventstore/bolt"
	"github.com/fiatjaf/khatru/policies"
	"github.com/fiatjaf/relay29"
	"github.com/joho/godotenv"
	"github.com/nbd-wtf/go-nostr"
	"golang.org/x/exp/slices"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	relayPrivateKey := os.Getenv("RELAY_PRIVATE_KEY")
	relayName := os.Getenv("RELAY_NAME")
	relayDescription := os.Getenv("RELAY_DESCRIPTION")
	relayContact := os.Getenv("RELAY_CONTACT")
	relayIconUrl := os.Getenv("RELAY_ICON_URL")

	relayPublicKey, err := nostr.GetPublicKey(relayPrivateKey)
	if err != nil {
		panic(err)
	}

	state := relay29.Init(relay29.Options{
		Domain:    "localhost:5577",
		DB:        &bolt.BoltBackend{Path: "./db"},
		SecretKey: relayPrivateKey,
	})

	if err := state.DB.Init(); err != nil {
		fmt.Print("hmm")
		panic(err)
	}

	// init relay
	state.Relay.Info.Name = relayName
	state.Relay.Info.Description = relayDescription
	state.Relay.Info.Contact = relayContact
	state.Relay.Info.Icon = relayIconUrl
	state.Relay.Info.PubKey = relayPublicKey
	state.Relay.Info.SupportedNIPs = append(state.Relay.Info.SupportedNIPs, 29)

	// extra policies
	state.Relay.RejectEvent = slices.Insert(state.Relay.RejectEvent, 0,
		policies.PreventLargeTags(64),
		policies.PreventTooManyIndexableTags(6, []int{9005}, nil),
		policies.RestrictToSpecifiedKinds(
			0, 9, 10, 11, 12, 10009,
			30023, 31922, 31923, 9802,
			9000, 9001, 9002, 9003, 9004, 9005, 9006, 9007,
			9021,
		),
		policies.PreventTimestampsInThePast(60),
		policies.PreventTimestampsInTheFuture(30),
	)

	// http routes
	state.Relay.Router().HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "nothing to see here, you must use a nip-29 powered client")
	})

	fmt.Println("running on http://0.0.0.0:5577")
	if err := http.ListenAndServe(":5577", state.Relay); err != nil {
		log.Fatal("failed to serve")
	}
}
