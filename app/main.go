package main

import (
	"github.com/cryptobase/scraper/model"
	"github.com/cryptobase/scraper/bitfinex"
	"github.com/cryptobase/scraper/bitstamp"
	"fmt"
	"log"
	"os"
	"encoding/csv"
	"strconv"
	//"time"
)

func main() {
	handler_wrapper(bitfinex.Scrape, "bitfinex")
	handler_wrapper(bitstamp.Scrape, "bitstamp")
}

func handler_wrapper(f func(int64) ([]model.Trade, error), name string) {
	existing, new, last_timestamp, err := handler(f, name)
	if err != nil {
		log.Printf("[%10s] Update %7s :: msg=[%s]", name, "failed", err)
	} else {
		log.Printf("[%10s] Update %7s :: msg=[#existing=%9d, last ts=%10d, #new=%5d]", name, "success", existing, last_timestamp, new)
	}
}

func handler(f func(int64) ([]model.Trade, error), name string) (int, int, int64, error) {
	output_file, err := prepare(name)
	if err != nil {
		return 0,0,0,err
	}

	//Load existing trades from file
	existing_trades, _ := LoadFromCsv(output_file)

	//Fetch last timestamp
	last_timestamp := int64(0)
	if len(existing_trades) > 0 {
		last_record := existing_trades[len(existing_trades)-1]
		last_timestamp = last_record.Timestamp
	}
	//log.Printf("Last trade timestamp: %d", last_timestamp)

	//Load new trades from api
	new_trades, err := f(last_timestamp)
	if err != nil {
		//log.Printf("Failed to load data")
		return 0,0,0,err
	}

	//Append new trades to file
	count, err1 := AppendToCsv(output_file, new_trades)
	if err1 != nil {
		return 0,0,0,err1
	}

	//log.Printf("Appended %d records", count)
	return len(existing_trades), count, last_timestamp, nil
}

func prepare(name string) (string, error) {
	path := "/Users/wilelb/crypto-scraper/"
	file := fmt.Sprintf("%s.csv", name)
	output_file := fmt.Sprintf("%s%s", path, file)

	_, err := os.Stat(path)
	if err != nil {
		err := os.MkdirAll(path, 0777)
		if err != nil {
			log.Fatal(err)
			return "", err
		} else {
			log.Printf("Created output directory: [%s]", path)
		}
	}

	return output_file, nil
}

func LoadFromCsv(file string) ([]model.Trade, error) {
	//log.Printf("Loading from file: [%s]", file)

	trades := []model.Trade{}

	csvfile, err := os.Open(file)
	if err != nil {
		return trades, err
	}
	defer csvfile.Close()

	reader := csv.NewReader(csvfile)
	reader.FieldsPerRecord = -1 // see the Reader struct information below

	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return trades, err
	}

	var trade model.Trade
	for _, record := range rawCSVdata {
		trade.Timestamp, _ = strconv.ParseInt(record[0], 10, 64)
		trade.TradeId, _ = strconv.ParseInt(record[1], 10, 64)
		trade.Exchange = record[2]
		trade.Type = record[3]
		trade.Amount, _ = strconv.ParseFloat(record[4], 64)
		trade.Price, _ = strconv.ParseFloat(record[5], 64)
		trades = append(trades, trade)
	}

	return trades, nil
}

func AppendToCsv(file string, trades []model.Trade) (int, error) {
	appended := 0
	csvfile, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return appended, err
	}
	defer csvfile.Close()

	writer := csv.NewWriter(csvfile)
	for _, trade := range trades {

		//tm := time.Unix(trade.Timestamp, 0)
		//fmt.Printf("%d-%d-%d\n", tm.Year(), tm.Month(), tm.Day())

		record := []string{
			fmt.Sprintf("%d", trade.Timestamp),
			fmt.Sprintf("%d", trade.TradeId),
			trade.Exchange,
			trade.Type,
			fmt.Sprintf("%.8f", trade.Amount),
			fmt.Sprintf("%.2f", trade.Price)}
		err := writer.Write(record)
		appended++
		if err != nil {
			return appended, err
		}
	}
	writer.Flush()

	return appended, nil
}