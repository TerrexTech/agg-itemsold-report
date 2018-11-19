package main

import (
	"log"
	"os"

	"github.com/TerrexTech/agg-itemsold-report/report"
	"github.com/TerrexTech/go-commonutils/commonutil"
	"github.com/TerrexTech/go-eventspoll/poll"

	"github.com/TerrexTech/go-kafkautils/kafka"
	tlog "github.com/TerrexTech/go-logtransport/log"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

var aggregateID int8 = 11

func validateEnv() error {
	missingVar, err := commonutil.ValidateEnv(
		"KAFKA_BROKERS",

		"KAFKA_CONSUMER_EVENT_GROUP",
		"KAFKA_CONSUMER_EVENT_QUERY_GROUP",

		"KAFKA_CONSUMER_EVENT_TOPIC",
		"KAFKA_CONSUMER_EVENT_QUERY_TOPIC",
		"KAFKA_PRODUCER_EVENT_QUERY_TOPIC",
		"KAFKA_PRODUCER_RESPONSE_TOPIC",

		"MONGO_HOSTS",
		"MONGO_DATABASE",
		"MONGO_AGG_COLLECTION",
		"MONGO_REPORT_COLLECTION",
		"MONGO_META_COLLECTION",

		"MONGO_CONNECTION_TIMEOUT_MS",
		"MONGO_RESOURCE_TIMEOUT_MS",
	)

	if err != nil {
		err = errors.Wrapf(
			err,
			"Env-var %s is required for testing, but is not set", missingVar,
		)
		return err
	}
	return nil
}

// func createData(numIterations int, repCollection *mongo.Collection) {
// 	newReport := []report.SoldItem{}
// 	for i := 0; i < numIterations; i++ {
// 		newReport = append(newReport, report.InsertItemSold())
// 	}

// 	for range newReport {
// 		for _, v := range newReport {
// 			insertResult, err := repCollection.InsertOne(v)
// 			if err != nil {
// 				err = errors.Wrap(err, "Unable to insert data")
// 				log.Println(err)
// 			}
// 			log.Println(insertResult)
// 		}
// 	}
// }

func main() {
	log.Println("Reading environment file")
	err := godotenv.Load("./.env")
	if err != nil {
		err = errors.Wrap(err,
			".env file not found, env-vars will be read as set in environment",
		)
		log.Println(err)
	}

	err = validateEnv()
	if err != nil {
		log.Fatalln(err)
	}

	aggCollection := os.Getenv("MONGO_AGG_COLLECTION")
	reportCollection := os.Getenv("MONGO_REPORT_COLLECTION")
	brokersStr := os.Getenv("KAFKA_BROKERS")
	brokers := *commonutil.ParseHosts(brokersStr)
	logTopic := os.Getenv("KAFKA_LOG_PRODUCER_TOPIC")
	serviceName := os.Getenv("SERVICE_NAME")

	log.Println("=================")
	log.Println(serviceName)

	prodConfig := &kafka.ProducerConfig{
		KafkaBrokers: brokers,
	}
	logger, err := tlog.Init(nil, serviceName, prodConfig, logTopic)
	if err != nil {
		err = errors.Wrap(err, "Error initializing Logger")
		log.Fatalln(err)
	}

	kc, err := loadKafkaConfig()
	if err != nil {
		err = errors.Wrap(err, "Error in KafkaConfig")
		logger.F(tlog.Entry{
			Description: err.Error(),
			ErrorCode:   1,
		}, kc)

	}

	//This is for report collection
	mc, err := loadMongoConfig(reportCollection, &report.SoldReport{})
	if err != nil {
		err = errors.Wrap(err, "Error in MongoConfig - trying to load SoldReport - mongoCollection")
		logger.F(tlog.Entry{
			Description: err.Error(),
			ErrorCode:   1,
		})
	}

	ioConfig := poll.IOConfig{
		ReadConfig: poll.ReadConfig{
			EnableQuery: true,
		},
		KafkaConfig: *kc,
		MongoConfig: *mc,
	}

	eventPoll, err := poll.Init(ioConfig)
	if err != nil {
		err = errors.Wrap(err, "Error creating EventPoll service")
		logger.F(tlog.Entry{
			Description: err.Error(),
			ErrorCode:   1,
		}, eventPoll)
	}

	client, err := CreateClient()
	if err != nil {
		err = errors.Wrap(err, "Error in MongoClient")
		logger.F(tlog.Entry{
			Description: err.Error(),
			ErrorCode:   1,
		})
	}

	itemSoldColl, err := CreateCollection(client, aggCollection, &report.SoldItem{})
	if err != nil {
		err = errors.Wrap(err, "Error in MongoCollection")
		logger.F(tlog.Entry{
			Description: err.Error(),
			ErrorCode:   1,
		})
	}

	for {
		select {
		case <-eventPoll.RoutinesCtx().Done():
			err = errors.New("service-context closed")
			logger.F(tlog.Entry{
				Description: err.Error(),
				ErrorCode:   1,
			})

		case eventResp := <-eventPoll.Query():
			go func(eventResp *poll.EventResponse) {
				if eventResp == nil {
					return
				}
				err := eventResp.Error
				if err != nil {
					err = errors.Wrap(err, "Error in Query-EventResponse")
					logger.E(tlog.Entry{
						Description: err.Error(),
						ErrorCode:   1,
					})
					return
				}
				kafkaResp := Query(logger, itemSoldColl, mc.AggCollection, &eventResp.Event)
				if kafkaResp != nil {
					eventPoll.ProduceResult() <- kafkaResp
				}
			}(eventResp)
		}
	}
}
