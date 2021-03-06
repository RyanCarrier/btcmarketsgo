package main

import (
	"fmt"

	"github.com/RyanCarrier/btcmarketsgo"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
)

var client *btcmarketsgo.BTCMarketsClient

func init() {

	var err error
	client, err = btcmarketsgo.NewDefaultClient(btcmarketsgo.GetKeys("api.secret"))
	log.SetLevel(log.DebugLevel)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	//got, err := client.GetOrderBook("BTC", "AUD")
	//got, err := btcmarketsgo.BTCMarketsClient{}.GetOrderBook("ETH", "BTC")
	//log.Info("Open orders output:")
	got, err := client.OrdersDetails(765500586)
	//got, err := client.GetOpenOrders("BCH", "AUD")
	print(got, err)
	//Ticker example
	/*quit := make(chan bool)
	client.Ticker(func(tr btcmarketsgo.TickResponse, err error) {
		fmt.Printf("%+v\n", tr)
	}, time.Second, quit)
	log.Info("quiting after 50 seconds")
	time.Sleep(time.Second * 5 * 10)
	quit <- true
	log.Info("quit")*/
}

func print(got interface{}, err error) {
	if err != nil {
		fmt.Println(err)
	}
	config := spew.NewDefaultConfig()
	config.Indent = "\t"
	config.Dump(got)
}
