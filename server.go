package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/julienschmidt/httprouter"
	log "gopkg.in/inconshreveable/log15.v2"
)

// Global variables
var (
	db            *localRDFStore
	cfg           *config
	l             = log.New()
	queues        Queues
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
	queues = append(queues, newQueue("addToIndex", 100, 2, newAddWorker))
	queues = append(queues, newQueue("rmFromIndex", 100, 1, newRmWorker))
	queues = append(queues, newQueue("syncToKoha", 1000, 1, newKohaSyncWorker))
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
	mux.POST("/queue/add", addToIndex)
	mux.POST("/queue/remove", rmFromIndex)
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
