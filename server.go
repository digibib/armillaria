package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/julienschmidt/httprouter"
	"github.com/knakk/sparql"
	log "gopkg.in/inconshreveable/log15.v2"
)

// Global variables
// TODO create env struct and move all globals into it
// env must be supplied to handlers or embedded in handler structs.
var (
	db            *localRDFStore
	cfg           *config
	l             = log.New()
	queues        Queues
	indexMappings map[string]map[string]string // indexMappings[profile]map[predicate] = property
	esIndexer     = Indexer{host: "http://" + os.Getenv("ELASTICSEARCH_PORT_9200_TCP_ADDR") + ":9200", client: http.DefaultClient}
	idGen         = newIdService()
	kohaCookies   http.CookieJar
	qBank         sparql.Bank
)

func main() {
	// Load configuration files
	var err error
	cfg, err = loadConfig("data/config.json")

	if err != nil {
		l.Error("failed to load config.json", log.Ctx{"error": err.Error()})
		os.Exit(1)
	}

	// Override with environment variables if present
	for _, env := range os.Environ() {
		key := strings.Split(env, "=")
		switch key[0] {
		case "SERVER_PORT":
			cfg.ServePort, _ = strconv.Atoi(key[1])
		case "VIRTUOSO_PORT_8890_TCP_ADDR":
			cfg.RDFStore.Endpoint = fmt.Sprintf("http://%s:8890/sparql-auth", key[1])
		case "SPARUL_USER":
			cfg.RDFStore.Username = key[1]
		case "SPARUL_PASS":
			cfg.RDFStore.Password = key[1]
		case "DEFAULT_GRAPH":
			cfg.RDFStore.DefaultGraph = key[1]
		case "ELASTICSEARCH_PORT_9200_TCP_ADDR":
			cfg.Elasticsearch = fmt.Sprintf("http://%s:9200/", key[1])
		}
	}

	// Setup logging
	l.SetHandler(log.StreamHandler(os.Stdout, log.TerminalFormat()))

	// Load index mappings (RDF predicate -> Elasticsearch document field)
	indexMappings, err = loadFromProfiles()
	if err != nil {
		l.Error("failed to load index mappings", log.Ctx{"error": err.Error()})
	}

	// Setup local repo
	db = newLocalRDFStore(
		cfg.RDFStore.Endpoint,
		cfg.RDFStore.Username,
		cfg.RDFStore.Password)

	// parse SPARQL queries
	qBank = sparql.LoadBank(bytes.NewBufferString(sparqlQueries))
	// Init idService
	res, err := db.Query(queryGetMax)
	if err != nil {
		l.Error("failed to get maximum Ids from RDF store; exiting", log.Ctx{"error": err.Error()})
		os.Exit(1) // Cannot continue without this information
	}
	err = idGen.Init(res)
	if err != nil {
		l.Error("failed to initalize idService; exiting", log.Ctx{"error": err.Error()})
		os.Exit(1) // Cannot continue without this information
	}

	// Initialize queues and workers
	queues = append(queues, newQueue("add", 100, 2, newAddWorker))
	queues = append(queues, newQueue("remove", 100, 1, newRmWorker))
	queues = append(queues, newQueue("KohaSync", 1000, 1, newKohaSyncWorker))
	queues.StartAll()

	// setup ElasticSearch proxy
	esHost, err := url.Parse(cfg.Elasticsearch)
	if err != nil {
		l.Error("unparsable Elasticsearch host address", log.Ctx{"error": err.Error()})
		os.Exit(1)
	}
	esProxy := httputil.NewSingleHostReverseProxy(esHost)

	// Routing
	mux := httprouter.New()
	mux.Handle("POST", "/search/*indexandtype", searchHandler(esProxy))
	mux.GET("/id/:type", getIdHandler)
	mux.GET("/rdf2marc", rdf2marcHandler)
	mux.POST("/resource", doResourceQuery)
	mux.POST("/external/:source", queryExternalSource)
	mux.HandlerFunc("GET", "/resource", serveFile("./data/html/resource.html"))
	mux.HandlerFunc("GET", "/", serveFile("./data/html/index.html"))
	mux.ServeFiles("/public/*filepath", http.Dir("./data/public/"))

	// Trap interutption signals to clean up before shutdown
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		<-interruptChan
		l.Info("interrupt signal received; exiting")

		// Stop queue workers, letting them complete any running tasks.
		queues.StopAll()

		os.Exit(0)
	}()

	// Start server
	l.Info("starting Armillaria server", log.Ctx{"port": cfg.ServePort})
	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.ServePort), mux)
	if err != nil {
		l.Error("http server crashed", log.Ctx{"details": err.Error()})
	}
}
