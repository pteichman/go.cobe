package ircbot

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	irc "github.com/fluffle/goirc/client"
	"github.com/pteichman/go.cobe"
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
		time.Sleep(backoffDuration(i))

		err := conn.Connect(o.Server)
		if err == nil {
			// The connection was successful.
			break
		}
	}
}

func RunForever(b *cobe.Brain, o *Options) {
	stop := make(chan bool)
	conn := irc.SimpleClient(o.Nick)
	conn.Me.Ident = o.Nick
	conn.Me.Name = o.Nick

	conn.AddHandler("connected", func(conn *irc.Conn, line *irc.Line) {
		for _, channel := range o.Channels {
			conn.Join(channel)
		}
	})

	conn.AddHandler("disconnected", func(conn *irc.Conn, line *irc.Line) {
		backoffConnect(conn, o)
	})

	// The space after comma/colon is needed so we won't treat
	// urls as messages spoken to http.
	userMsg := regexp.MustCompile(`^(\S+)[,:]\s(.*?)$`)

	conn.AddHandler("privmsg", func(conn *irc.Conn, line *irc.Line) {
		user := line.Nick
		if in(o.Ignore, line.Nick) {
			return
		}

		target := line.Args[0]
		if !in(o.Channels, target) {

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

		b.Learn(msg)

		if to == o.Nick {
			reply := b.Reply(msg)
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
