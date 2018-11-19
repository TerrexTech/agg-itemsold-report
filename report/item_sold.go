package report

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/TerrexTech/go-mongoutils/mongo"
	"github.com/mongodb/mongo-go-driver/bson"
	mgo "github.com/mongodb/mongo-go-driver/mongo"
	"github.com/pkg/errors"
)

func ItemSoldReport(aggParams SoldItemParams, itemSoldColl *mongo.Collection) ([]interface{}, error) {

	if aggParams.Timestamp.Lt == 0 || aggParams.Timestamp.Gt == 0 {
		err := errors.New("Missing timestamp value")
		log.Println(err)
		return nil, err
	}
	input, err := json.Marshal(aggParams)
	if err != nil {
		err = errors.Wrap(err, "Unable to marshal aggParams")
		log.Println(err)
		return nil, err
	}

	log.Println(input)
	log.Println(aggParams)

	pipelineBuilder := fmt.Sprintf(`[
		{
			"$match": %s
		},
		{
			"$group" : {
			"_id" : {"sku" : "$sku","name":"$name"},
			"avg_sold": {
				"$avg": "$weight",
			},
			"avg_total": {
				"$avg": "$totalWeight",
			}
		}
		}
	]`, input)

	pipelineAgg, err := bson.ParseExtJSONArray(pipelineBuilder)
	if err != nil {
		err = errors.Wrap(err, "Query: Error in generating pipeline for report")
		log.Println(err)
		return nil, err
	}

	findResult, err := itemSoldColl.Aggregate(pipelineAgg)
	if err != nil {
		err = errors.Wrap(err, "Query: Error in getting aggregate results ")
		log.Println(err)
		return nil, err
	}
	return findResult, nil
}

func CreateReport(reportGen SoldReport, reportColl *mongo.Collection) (*mgo.InsertOneResult, error) {
	insertRep, err := reportColl.InsertOne(reportGen)
	if err != nil {
		err = errors.Wrap(err, "Query: Error in generating report ")
		log.Println(err)
		return nil, err
	}
	return insertRep, nil
}
