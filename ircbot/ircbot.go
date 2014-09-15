package ircbot

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	irc "github.com/fluffle/goirc/client"
	cobe "github.com/pteichman/go.cobe"
)

type Options struct {
	Server   string
	Nick     string
	Channels []string
	Ignore   []string
}

// Backoff policy, milliseconds per attempt. End up with 30s attempts.
var backoff = []int{0, 0, 10, 30, 100, 300, 1000, 3000, 10000, 30000}

func backoffDuration(i int) time.Duration {
	if i < len(backoff) {
		return time.Duration(backoff[i]) * time.Millisecond
	}

	return time.Duration(backoff[len(backoff)-1]) * time.Millisecond
}

func backoffConnect(conn *irc.Conn, o *Options) {
	for i := 0; true; i++ {
		wait := backoffDuration(i)
		time.Sleep(wait)

		err := conn.Connect()
		if err == nil {
			// The connection was successful.
			break
		}

		log.Printf("Connection to %s failed: %s [%dms]", o.Server, err,
			int64(wait/time.Millisecond))
	}
}

func RunForever(b *cobe.Cobe2Brain, o *Options) {
	stop := make(chan bool)
	conn := irc.SimpleClient(o.Nick)

	conn.Config().Server = o.Server

	conn.Me().Ident = o.Nick
	conn.Me().Name = o.Nick

	conn.HandleFunc("connected", func(conn *irc.Conn, line *irc.Line) {
		log.Printf("Connected to %s. Joining %s.", o.Server,
			strings.Join(o.Channels, ", "))
		for _, channel := range o.Channels {
			conn.Join(channel)
		}
	})

	conn.HandleFunc("disconnected", func(conn *irc.Conn, line *irc.Line) {
		log.Printf("Disconnected from %s.", o.Server)
		backoffConnect(conn, o)
	})

	conn.HandleFunc("kick", func(conn *irc.Conn, line *irc.Line) {
		if line.Args[1] == o.Nick {
			var channel = line.Args[0]
			log.Printf("Kicked from %s. Rejoining.", channel)
			conn.Join(channel)
		}
	})

	// The space after comma/colon is needed so we won't treat
	// urls as messages spoken to http.
	userMsg := regexp.MustCompile(`^(\S+)[,:]\s(.*?)$`)

	conn.HandleFunc("privmsg", func(conn *irc.Conn, line *irc.Line) {
		user := line.Nick
		if in(o.Ignore, user) {
			log.Printf("Ignoring privmsg from %s", user)
			return
		}

		target := line.Args[0]
		if !in(o.Channels, target) {
			log.Printf("Ignoring privmsg on %s", target)
			return
		}

		var to, msg string

		groups := userMsg.FindStringSubmatch(line.Args[1])
		if len(groups) > 0 {
			to = groups[1]
			msg = groups[2]
		} else {
			msg = line.Args[1]
		}

		msg = strings.TrimSpace(msg)

		log.Printf("Learn: %s", msg)
		b.Learn(msg)

		if to == o.Nick {
			reply := b.Reply(msg)
			log.Printf("Reply: %s", reply)
			conn.Privmsg(target, fmt.Sprintf("%s: %s", user, reply))
		}
	})

	backoffConnect(conn, o)
	<-stop
}

func in(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}

	return false
}
