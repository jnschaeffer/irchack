package main

import (
	"github.com/jnschaeffer/irchack/nethack"
	"github.com/jnschaeffer/irchack/irc"
	"log"
)

func main() {
	nh := nethack.NewNetHack("/tmp/nethack.sock", "color", "statushilites:10")

	defer nh.Close()

	errStart := nh.Start()
	if errStart != nil {
		log.Printf("error during start: %s", errStart.Error())
		return
	}

	handler := irc.NewTeeHandler(nh)

	client := irc.NewClient("ircnethackbot", "irc.freenode.net", "", "#ufosolutions", true)

	defer client.Close()

	client.RegisterHandler("PRIVMSG", handler)

	errConnect := client.Connect()
	if errConnect != nil {
		log.Printf("error during connect: %s", errConnect.Error())

		return
	}

	errWait := nh.Wait()

	if errWait != nil {
		log.Printf("error during wait: %s", errWait.Error())
	}

	return
}
