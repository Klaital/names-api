package main

import (
	"flag"
	"github.com/klaital/names-api/pkg/config"
	"github.com/klaital/names-api/pkg/people"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

func main() {
	cfg := config.ServiceConfig{
		LogLevel:       "debug",
		LogFormat:      "text",
		DatabaseHost:   "",
		DatabaseName:   "af_names",
		DatabasePort:   3306,
		DatabaseUser:   "af_namer",
		DatabasePass:   "",
		DatabaseDriver: "mysql",
	}
	var peopleFile string
	var envFile string

	flag.StringVar(&cfg.LogLevel, "loglevel", "debug", "Set the log level")
	flag.StringVar(&cfg.LogFormat, "logformat", "text", "Set the logrus formatter. 'text', 'json', 'prettyjson'")
	flag.StringVar(&peopleFile, "people", "", "Upload this file of people names")
	flag.StringVar(&envFile, "dbconf", "db.env", "Use this env file to load DB connection config")
	flag.Parse()

	logger := cfg.GetLogger()
	envBytes, err := ioutil.ReadFile(envFile)
	if err != nil {
		logger.WithError(err).Fatal("Failed to read env file")
	}
	envConf := parseEnvFile(string(envBytes))
	cfg.DatabaseHost = envConf["DB_HOST"]
	cfg.DatabaseName = envConf["DB_NAME"]
	cfg.DatabaseUser = envConf["DB_USER"]
	cfg.DatabasePass = envConf["DB_PASS"]
	db, err := cfg.GetDbConn()
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}

	// Load the data file with people data
	peopleBytes, err := ioutil.ReadFile(peopleFile)
	if err != nil {
		logger.WithError(err).Fatal("Failed to read people file")
	}
	var newNameData people.NameFile
	err = yaml.Unmarshal(peopleBytes, &newNameData)
	if err != nil {
		logger.WithError(err).Fatal("Failed to unmarshal people file")
	}

	// Load the existing people in the database
	dbPeople, err := people.LoadAllPeople(db)
	if err != nil {
		logger.WithError(err).Fatal("Failed to load people from the database")
	}

	// Add new people to the database
	for _, person := range newNameData.Names {
		existingPerson := false
		for _, dbPerson := range dbPeople {
			if dbPerson.Name == person.Name {
				existingPerson = true
				// TODO: update existing entries
				break
			}
		}
		if !existingPerson {
			err = person.Insert(db)
			if err != nil {
				logger.WithError(err).WithField("person", person).Error("Failed to insert new person")
			} else {
				logger.WithField("name", person.Name).Info("Uploaded new person")
			}
		}
	}

	// TODO: delete people removed from the list?
}

func parseEnvFile(contents string) map[string]string {
	envMap := make(map[string]string, 0)
	lines := strings.Split(contents, "\n")
	for _, line := range lines {
		tokens := strings.SplitN(line, "=", 2)
		if len(tokens) < 2 {
			continue
		}
		k := strings.TrimSpace(tokens[0])
		v := strings.TrimSpace(tokens[1])
		envMap[k] = v
	}

	return envMap
}
