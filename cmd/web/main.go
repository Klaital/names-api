package main

import (
	"github.com/emicklei/go-restful"
	"github.com/klaital/names-api/pkg/config"
	"github.com/klaital/names-api/pkg/people"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)
func main() {
	cfg := config.LoadConfig()
	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		logLevel = log.DebugLevel
	}
	log.SetLevel(logLevel)
	switch cfg.LogFormat {
	case "prettyjson":
		log.SetFormatter(&log.JSONFormatter{
			PrettyPrint:       true,
		})
	case "json":
		log.SetFormatter(&log.JSONFormatter{
			PrettyPrint:       false,
		})
	default:
		log.SetFormatter(&log.TextFormatter{})
	}
	log.Info("Configuring webserver")

	container := restful.NewContainer()
	service := new(restful.WebService)
	service.Path("/names").ApiVersion("1.0.0").Doc("Fetch interesting names")

	// People
	service.Route(
		service.GET("/people").
			To(FindPeopleHandler))
	container.Add(service)
	// TODO: places

	httpServer := &http.Server{
		Addr:              ":8080",
		Handler:           container,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
	log.Infof("Server listening on %s", cfg.Port)
	log.Fatal(httpServer.ListenAndServe())
}

func FindPeopleHandler(request *restful.Request, response *restful.Response) {
	cfg := config.LoadConfig()
	db, err := cfg.GetDbConn()
	if err != nil {
		log.WithError(err).Error("Failed to connect to db")
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	peopleSet, err := people.LoadAllPeople(db)
	if err != nil {
		log.WithError(err).Error("Failed to query for people")
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = response.WriteEntity(peopleSet)
	if err != nil {
		log.WithError(err).Error("Failed to serialize payload")
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
}
