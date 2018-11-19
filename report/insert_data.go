package report

import (
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/TerrexTech/uuuid"
	"github.com/pkg/errors"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func random(min, max int64) int64 {
	return rand.Int63n(max-min) + min
}

func generateRandomValue(num1, num2 int64) int64 {
	// rand.Seed(time.Now().Unix())
	return random(num1, num2)
}

func GenFakeBarcode(barType string) int64 {
	var num int64
	if barType == "upc" {
		num = generateRandomValue(111111111111, 999999999999)
	}
	if barType == "sku" {
		num = generateRandomValue(11111111, 99999999)
	}
	return num
}

func generateNewUUID() uuuid.UUID {
	uuid, err := uuuid.NewV4()
	if err != nil {
		err = errors.Wrap(err, "Unable to generate UUID")
		log.Println(err)
	}
	return uuid
}

var productsName = []string{"Banana", "Orange", "Apple", "Mango", "Strawberry", "Tomato", "Lettuce", "Pear", "Grapes", "Sweet Pepper"}
var lot = []string{"A101", "B201", "O301", "M401", "S501", "T601", "L701", "P801", "G901", "SW1001"}

func InsertItemSold() SoldItem {
	randNameAndLocation := generateRandomValue(1, 10)
	randTotalWeight := generateRandomValue(100, 300)
	randSoldWeight := generateRandomValue(1, randTotalWeight)
	name := productsName[randNameAndLocation-1]
	lot := lot[randNameAndLocation]
	sku := GenFakeBarcode("sku")
	randDateArr := generateRandomValue(-50, 6)
	timestamp := time.Now().AddDate(0, 0, int(randDateArr)).Unix()

	t := strconv.Itoa(int(sku))

	item := SoldItem{
		ItemID:      generateNewUUID(),
		SaleID:      generateNewUUID(),
		SKU:         t,
		Name:        name,
		Lot:         lot,
		TotalWeight: float64(randTotalWeight),
		Weight:      float64(randSoldWeight),
		Timestamp:   timestamp,
	}

	return item
}
