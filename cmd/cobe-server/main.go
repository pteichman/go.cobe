package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"github.com/pteichman/go.cobe"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var (
	brains     cobe.BrainPool
	clog            = logging.MustGetLogger("cobe.server")
	isLearning bool = false
)

// HTTP REST interface to cobe. Can handle multiple connections to brains.
func runForever(cp cobe.BrainPool, addr string, learn bool) {
	clog.Debug("Starting cobe REST HTTP server listening on %s.", addr)

	brains = cp
	isLearning = learn

	router := mux.NewRouter()
	router.HandleFunc("/{brain}/reply", replyHandler)
	router.HandleFunc("/", notFoundHandler)

	http.Handle("/", router)
	http.ListenAndServe(addr, nil)
}

// Reply handler. POST questions in plaintext to this endpoint
// and get an answer in plaintext back.
func replyHandler(w http.ResponseWriter, r *http.Request) {
	b, err := brains.Get(mux.Vars(r)["brain"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	in := strings.TrimSpace(string(body))
	out := b.Reply(in)

	fmt.Fprint(w, out)

	if isLearning {
		b.Learn(in)
	}
}

// Handles 404s. Does not return any text in body as this
// can be mistaken to be a valid answer.
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

// Implements a minimalistic REST HTTP server.
//
// When started with learning enabled will learn from input text
// otherwise will be in read only mode.
//
// Currently the only endpoint avaiable is '/<brain>/reply'. To get a
// reply POST plaintext to this endpoint. Replies are just plaintext,
// too.
//
// One or multiple brains must be specified when starting the server.
// This protects against arbitrary creation of brains while also
// allowing to switch brains on a per request basis (i.e. to support
// multiple languages).
func main() {
	app := cli.NewApp()
	app.Name = "cobe-server"
	app.Usage = "minimal REST HTTP server"
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "learn, L", Usage: "enables learning from input"},
		cli.StringSliceFlag{Name: "brain, b", Value: &cli.StringSlice{}, Usage: "name of brain to use; repeat for multiple brains."},
	}
	app.Action = func(c *cli.Context) {
		if len(c.Args()) < 1 {
			clog.Fatal("Usage: server <host:[port]>")
		}
		// Populate brain pool.
		brains := cobe.NewBrainPool()

		for _, v := range c.StringSlice("brain") {
			if err := brains.Add(v); err != nil {
				clog.Fatal(err)
			}
		}
		runForever(brains, c.Args()[0], c.Bool("learn"))
	}
	app.Run(os.Args)
}
