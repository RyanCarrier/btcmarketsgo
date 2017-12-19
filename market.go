package btcmarketsgo

import (
	"encoding/json"
	"errors"
	"time"

	ccg "github.com/RyanCarrier/cryptoclientgo"
)

//TickResponse is the response recieved when requesting market tick
type TickResponse struct {
	BestBid    float64
	BestAsk    float64
	LastPrice  float64
	Currency   string
	Instrument string
	Timestamp  int64
	Volume24h  float64
}

//DefaultTick get current tick details
func (c BTCMarketsClient) DefaultTick() (tr TickResponse, err error) {
	all, err := getBody(c.Domain + "/market/" + "BTC" + "/" + "AUD" + "/tick")
	if err != nil {
		return
	}
	err = json.Unmarshal(all, &tr)
	return
}

//Tick get current tick details
func (c BTCMarketsClient) Tick(CurrencyFrom, CurrencyTo string) (tr ccg.TickFloat, err error) {
	all, err := getBody(c.Domain + "/market/" + CurrencyTo + "/" + CurrencyFrom + "/tick")
	if err != nil {
		return
	}
	err = json.Unmarshal(all, &tr)
	return
}

//OrderBookResponse is the response recieved when requesting the order book
type OrderBookResponse struct {
	Currency   string
	Instrument string
	Timestamp  int64
	Asks       [][]float64
	Bids       [][]float64
}

func (obr OrderBookResponse) convert() (ccg.OrderBook, error) {
	result := ccg.OrderBook{
		PrimaryCurrency:   obr.Instrument,
		SecondaryCurrency: obr.Currency,
		BuyOrders:         ccg.Orders(make([]ccg.Order, len(obr.Bids))),
		SellOrders:        ccg.Orders(make([]ccg.Order, len(obr.Asks))),
	}
	for i, b := range obr.Bids {
		if len(b) != 2 {
			return ccg.OrderBook{}, errors.New("Bid not correct size")
		}
		result.BuyOrders[i].Price = ccg.ConvertFromFloat(b[0])
		result.BuyOrders[i].Volume = ccg.ConvertFromFloat(b[1])
	}
	for i, a := range obr.Asks {
		if len(a) != 2 {
			return ccg.OrderBook{}, errors.New("Ask not correct size")
		}
		result.SellOrders[i].Price = ccg.ConvertFromFloat(a[0])
		result.SellOrders[i].Volume = ccg.ConvertFromFloat(a[1])
	}
	return result, nil
}

//TradeResponse is a single response of a trade
type TradeResponse struct {
	Tid    int64
	Amount float64
	Price  float64
	Date   int64
}

//TradesResponse is the trades returned from a trades request
type TradesResponse []TradeResponse

func (tr TradesResponse) convert(max int) ccg.RecentTrades {
	result := ccg.RecentTrades{}
	result.Timestamp = time.Now()
	result.Trades = ccg.Trades(make([]ccg.Trade, min(len(tr), max)))
	for i, t := range tr {
		if i >= max {
			break
		}
		result.Trades[i].Amount = ccg.ConvertFromFloat(t.Amount)
		result.Trades[i].Price = ccg.ConvertFromFloat(t.Price)
	}
	return result
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

//GetOrderBook gets the orders for the relevant currencies
func (c BTCMarketsClient) GetOrderBook(PrimaryCurrency, SecondaryCurrency string) (ccg.OrderBook, error) {
	all, err := getBody(c.Domain + "/market/" + PrimaryCurrency + "/" + SecondaryCurrency + "/orderbook")
	if err != nil {
		return ccg.OrderBook{}, err
	}
	var obr OrderBookResponse
	err = json.Unmarshal(all, &obr)
	if err != nil {
		return ccg.OrderBook{}, err
	}
	return obr.convert()
}

//GetRecentTrades gets most recent trades limited by historyAmount
func (c BTCMarketsClient) GetRecentTrades(PrimaryCurrency, SecondaryCurrency string, historyAmount int) (ccg.RecentTrades, error) {
	var all []byte
	var err error
	all, err = getBody(c.Domain + "/market/" + PrimaryCurrency + "/" + SecondaryCurrency + "/trades")
	if err != nil {
		return ccg.RecentTrades{}, err
	}
	var tr TradesResponse
	err = json.Unmarshal(all, &tr)
	if err != nil {
		return ccg.RecentTrades{}, err
	}
	result := tr.convert(historyAmount)
	result.PrimaryCurrency = PrimaryCurrency
	result.SecondaryCurrency = SecondaryCurrency
	return result, nil
}
