package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	irc "github.com/fluffle/goirc/client"
	logging "github.com/op/go-logging"
	"github.com/pteichman/go.cobe"
)

type Options struct {
	Server   string
	Nick     string
	Channels []string
	Ignore   []string
}

var clog = logging.MustGetLogger("cobe.ircbot")

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

		err := conn.Connect(o.Server)
		if err == nil {
			// The connection was successful.
			break
		}

		clog.Warning("Connection to %s failed: %s [%dms]", o.Server, err,
			int64(wait/time.Millisecond))
	}
}

func runForever(b *cobe.Cobe2Brain, o *Options) {
	stop := make(chan bool)
	conn := irc.SimpleClient(o.Nick)
	conn.Me.Ident = o.Nick
	conn.Me.Name = o.Nick

	conn.AddHandler("connected", func(conn *irc.Conn, line *irc.Line) {
		clog.Notice("Connected to %s. Joining %s.", o.Server,
			strings.Join(o.Channels, ", "))
		for _, channel := range o.Channels {
			conn.Join(channel)
		}
	})

	conn.AddHandler("disconnected", func(conn *irc.Conn, line *irc.Line) {
		clog.Warning("Disconnected from %s.", o.Server)
		backoffConnect(conn, o)
	})

	conn.AddHandler("kick", func(conn *irc.Conn, line *irc.Line) {
		if line.Args[1] == o.Nick {
			var channel = line.Args[0]
			clog.Notice("Kicked from %s. Rejoining.", channel)
			conn.Join(channel)
		}
	})

	// The space after comma/colon is needed so we won't treat
	// urls as messages spoken to http.
	userMsg := regexp.MustCompile(`^(\S+)[,:]\s(.*?)$`)

	conn.AddHandler("privmsg", func(conn *irc.Conn, line *irc.Line) {
		user := line.Nick
		if in(o.Ignore, user) {
			clog.Debug("Ignoring privmsg from %s", user)
			return
		}

		target := line.Args[0]
		if !in(o.Channels, target) {
			clog.Debug("Ignoring privmsg on %s", target)
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

func main() {
	app := cli.NewApp()
	app.Name = "cobe-ircbot"
	app.Usage = "IRC bot"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "server", Usage: "IRC server to connect to (host:port)"},
		cli.StringFlag{Name: "channels", Value: "#cobe", Usage: "IRC channel/s to join"},
		cli.StringFlag{Name: "nick", Value: "cobe", Usage: "nickname of the bot"},
		cli.StringFlag{Name: "brain, b", Value: "cobe.brain", Usage: "name of the sqlite file to use"},
	}
	app.Action = func(c *cli.Context) {
		brain, err := cobe.OpenCobe2Brain(c.String("brain"))
		if err != nil {
			clog.Fatal(err)
		}
		opts := &Options{
			c.String("server"),
			c.String("nick"),
			[]string{c.String("channels")},
			nil,
		}
		runForever(brain, opts)
	}
	app.Run(os.Args)
}
