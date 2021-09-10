package main

import (
	"fmt"
	"github.com/djumanoff/amqp"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	setdata_common "github.com/kirigaikabuto/setdata-common"
	"github.com/kirigaikabuto/setdata_questionnaire"
	"github.com/urfave/cli"
	"net/http"
	"os"
	"strconv"
)

var (
	configPath           = ".env"
	version              = "0.0.0"
	amqpHost             = ""
	amqpPort             = 0
	postgresUser         = ""
	postgresPassword     = ""
	postgresDatabaseName = ""
	postgresHost         = ""
	postgresPort         = 5432
	postgresParams       = ""
	flags                = []cli.Flag{
		&cli.StringFlag{
			Name:        "config, c",
			Usage:       "path to .env config file",
			Destination: &configPath,
		},
	}
)

func parseEnvFile() {
	// Parse config file (.env) if path to it specified and populate env vars
	if configPath != "" {
		godotenv.Overload(configPath)
	}
	amqpHost = os.Getenv("RABBIT_HOST")
	amqpPortStr := os.Getenv("RABBIT_PORT")
	amqpPort, _ = strconv.Atoi(amqpPortStr)
	if amqpPort == 0 {
		amqpPort = 5672
	}
	if amqpHost == "" {
		amqpHost = "localhost"
	}
	postgresUser = os.Getenv("POSTGRES_USER")
	postgresPassword = os.Getenv("POSTGRES_PASSWORD")
	postgresDatabaseName = os.Getenv("POSTGRES_DATABASE")
	postgresParams = os.Getenv("POSTGRES_PARAMS")
	portStr := os.Getenv("POSTGRES_PORT")
	postgresPort, _ = strconv.Atoi(portStr)
	postgresHost = os.Getenv("POSTGRES_HOST")
	if postgresHost == "" {
		postgresHost = "localhost"
	}
	if postgresPort == 0 {
		postgresPort = 5432
	}
	if postgresUser == "" {
		postgresUser = "setdatauser"
	}
	if postgresPassword == "" {
		postgresPassword = "123456789"
	}
	if postgresDatabaseName == "" {
		postgresDatabaseName = "setdata"
	}
	if postgresParams == "" {
		postgresParams = "sslmode=disable"
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "Set data questions api"
	app.Description = ""
	app.Usage = "set data run"
	app.UsageText = "set data run"
	app.Version = version
	app.Flags = flags
	app.Action = run

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

func run(c *cli.Context) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	parseEnvFile()
	amqpConfig := amqp.Config{
		AMQPUrl: "amqps://futohrkk:Qq4imfTpgcDawG6bzuSnJALRg-a6xqZl@toad.rmq.cloudamqp.com/futohrkk",
	}
	sess := amqp.NewSession(amqpConfig)
	err := sess.Connect()
	if err != nil {
		return err
	}
	clt, err := sess.Client(amqp.ClientConfig{})
	if err != nil {
		return err
	}
	amqpRequests := setdata_questionnaire.NewAmqpRequests(clt)
	service := setdata_questionnaire.NewQuestionsService(amqpRequests)
	httpEndpoints := setdata_questionnaire.NewHttpEndpoints(setdata_common.NewCommandHandler(service))
	router := mux.NewRouter()

	router.Methods("POST").Path("/questions").HandlerFunc(httpEndpoints.MakeCreateQuestionEndpoint())
	router.Methods("PUT").Path("/questions").HandlerFunc(httpEndpoints.MakeUpdateQuestionEndpoint())
	router.Methods("GET").Path("/questions").HandlerFunc(httpEndpoints.MakeListQuestionEndpoint())
	router.Methods("PUT").Path("/questions/add").HandlerFunc(httpEndpoints.MakeAddFieldToQuestionEndpoint())
	router.Methods("GET").Path("/questions/{name}").HandlerFunc(httpEndpoints.MakeGetQuestionsByQuestionnaireName("name"))

	router.Methods("POST").Path("/questionnaire").HandlerFunc(httpEndpoints.MakeCreateQuestionnaireEndpoint())
	router.Methods("GET").Path("/questionnaire").HandlerFunc(httpEndpoints.MakeListQuestionnaireEndpoint())
	router.Methods("GET").Path("/questionnaire/{name}").HandlerFunc(httpEndpoints.MakeGetQuestionnaireByName("name"))
	router.Methods("PUT").Path("/questionnaire/add").HandlerFunc(httpEndpoints.MakeAddQuestionToQuestionnaireEndpoint())
	router.Methods("PUT").Path("/questionnaire/remove").HandlerFunc(httpEndpoints.MakeDeleteQuestionFromQuestionnaireEndpoint())

	router.Methods("POST").Path("/order").HandlerFunc(httpEndpoints.MakeCreateOrderEndpoint())
	router.Methods("GET").Path("/order/{name}").HandlerFunc(httpEndpoints.MakeListOrderEndpoint("name"))
	router.Methods("POST").Path("/order/consultation").HandlerFunc(httpEndpoints.MakeSendOrderForConsultation())
	router.Methods("POST").Path("/order/email").HandlerFunc(httpEndpoints.MakeSendOrderEmail())
	http.ListenAndServe(":"+port, router)
	return nil
}
