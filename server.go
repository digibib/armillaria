package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/julienschmidt/httprouter"
	log "gopkg.in/inconshreveable/log15.v2"
)

// Global variables and constants
var (
	db            *localRDFStore
	cfg           *config
	l             = log.New()
	queueAdd      Queue
	queueRemove   Queue
	queueKohaSync Queue
	indexMappings map[string]map[string]string // indexMappings[profile]map[predicate] = property
	esIndexer     = Indexer{host: "http://localhost:9200", client: http.DefaultClient}
	idGen         = newIdService()
	kohaCookies   http.CookieJar
)

func main() {
	// Load configuration files
	var err error
	cfg, err = loadConfig("data/config.json")
	if err != nil {
		l.Error("failed to load config.json", log.Ctx{"error": err.Error()})
		os.Exit(1)
	}

	// Setup logging
	l.SetHandler(log.MultiHandler(
		log.LvlFilterHandler(log.LvlInfo, log.Must.FileHandler(cfg.LogFile, log.LogfmtFormat())),
		log.StreamHandler(os.Stdout, log.TerminalFormat())),
	)

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
	queueAdd = newQueue("addToIndex", 100, 2, newAddWorker)
	go queueAdd.runDispatcher()
	queueRemove = newQueue("rmFromIndex", 100, 1, newRmWorker)
	go queueRemove.runDispatcher()
	queueKohaSync = newQueue("syncToKoha", 1000, 1, newKohaSyncWorker)
	go queueKohaSync.runDispatcher()

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
	mux.POST("/RDF/resource", doResourceQuery)
	mux.POST("/queue/add", addToIndex)
	mux.POST("/queue/remove", rmFromIndex)
	mux.POST("/external/:source", queryExternalSource)
	mux.HandlerFunc("GET", "/resource", serveFile("./data/html/resource.html"))
	mux.HandlerFunc("GET", "/", serveFile("./data/html/index.html"))
	mux.ServeFiles("/public/*filepath", http.Dir("./data/public/"))

	// Start server
	l.Info("starting Armillaria server", log.Ctx{"port": cfg.ServePort})
	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.ServePort), mux)
	if err != nil {
		l.Error("http server crashed", log.Ctx{"details": err.Error()})
	}
}
