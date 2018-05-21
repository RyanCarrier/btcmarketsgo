package btcmarketsgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	ccg "github.com/RyanCarrier/cryptoclientgo"

	"github.com/davecgh/go-spew/spew"
)

const btcMin = int64(100000)
const multiplier = int64(100000000)

//OrderStatuses is the current available order statuses
//const OrderStatuses = []string{"New", "Placed", "Failed", "Error", "Cancelled",	"Partially Canceled", "Fully Matched", "Partially Matched"}

//OrderRequest is an order request struct for parsing into json
type OrderRequest struct {
	CurrencySecondary string `json:"currency"`
	CurrencyPrimary   string `json:"instrument"`
	Price             int64  `json:"price"`
	Volume            int64  `json:"volume"`
	OrderSide         string `json:"orderSide"`       //Camel case
	OrderType         string `json:"ordertype"`       //the lowercase T is important...
	ClientRequestID   string `json:"clientRequestId"` //Camel case
}

//OrderResponse is the response from submitting an order
type OrderResponse struct {
	Success         bool
	ErrorCode       int
	ErrorMessage    string
	ID              int
	ClientRequestID string
}

func (or OrderResponse) convert() ccg.PlacedOrder {
	return ccg.PlacedOrder{OrderID: or.ID}
}

//CancelOrdersRequest is the struct used to request the cancelation of an order(s)
type CancelOrdersRequest struct {
	OrderIds []int `json:"orderIds"`
}

//CancelOrdersResponse is the response received when canceling an order(s)
type CancelOrdersResponse struct {
	Success      bool
	ErrorCode    int
	ErrorMessage string
	Responses    []CancelOrderResponse
}

//CancelOrderResponse is the individual order cancelation response
type CancelOrderResponse struct {
	Success      bool
	ErrorCode    int
	ErrorMessage string
	ID           int
}

//CreateOrder creates an order at specified price and volume
func (c BTCMarketsClient) createOrder(CurrencyPrimary, CurrencySecondary string,
	Price, Volume int64, Buy bool, Market bool) (ccg.PlacedOrder, error) {
	if Volume < btcMin {
		return ccg.PlacedOrder{}, errors.New(
			fmt.Sprint("Volume must be ", btcMin, " minimum (", strconv.FormatFloat(
				float64(btcMin)/float64(multiplier), 'f', 3, 64)+"BTC)",
			),
		)
	}
	URI := "/order/create"
	or := OrderRequest{
		CurrencyPrimary:   CurrencyPrimary,
		CurrencySecondary: CurrencySecondary,
		Price:             Price,
		Volume:            Volume,
		ClientRequestID:   "1",
	}
	if Buy {
		or.OrderSide = "Bid"
	} else {
		or.OrderSide = "Ask"
	}
	if Market {
		or.OrderType = "Market"
	} else {
		or.OrderType = "Limit"
	}
	got, err := c.signAndPost(URI, or)
	if err != nil {
		err = errors.New("Error signing and POST'ing;" + err.Error() + "\n" + string(got))
	}
	var orderR OrderResponse
	err = json.Unmarshal(got, &orderR)
	if err != nil {
		err = errors.New("Error unmarshaling response;" + err.Error() + "\n" + string(got))
	}
	if !orderR.Success {
		return ccg.PlacedOrder{}, errors.New("Order failed; " + orderR.ErrorMessage + "; Could potentially not have the funds asked, or currencies incorrect")
	}
	if err != nil {
		return orderR.convert(), errors.New("Order failed, but response was success;" + err.Error())
	}
	return orderR.convert(), err
}

//CancelOrders requests the cancelation of an order(s)
func (c BTCMarketsClient) CancelOrders(orderIDs ...int) (CancelOrdersResponse, error) {
	URI := "/order/cancel"
	cor := CancelOrdersRequest{OrderIds: orderIDs}
	got, err := c.signAndPost(URI, cor)
	var cancelOR CancelOrdersResponse
	err = json.Unmarshal(got, &cancelOR)
	if err != nil {
		err = errors.New("Error unmarshaling response;" + err.Error() + "\n" + string(got))
	}
	return cancelOR, err
}

//CancelOrder cancels a single order
func (c BTCMarketsClient) CancelOrder(orderID int) error {
	got, err := c.CancelOrders(orderID)
	if err != nil {
		return err
	}
	if !got.Success {
		return errors.New("Cancel was not successful; " + got.ErrorMessage)
	}
	return nil
}

//OrderHistoryRequest gets the users order history
type OrderHistoryRequest struct {
	SecondaryCurrency string `json:"currency"`
	PrimaryCurrency   string `json:"instrument"`
	Limit             int    `json:"limit"`
	Since             int64  `json:"since"`
}

//OrderHistoryResponse is the response returned when requesting the history of a user
type OrderHistoryResponse struct {
	Success      bool
	ErrorCode    int
	ErrorMessage string
	Orders       []OrderHistorySingleResponse
}

func (ohr OrderHistoryResponse) convert() ccg.OrdersDetails {
	result := make([]ccg.OrderDetails, len(ohr.Orders))
	for i, o := range ohr.Orders {
		result[i] = o.convert()
	}
	return ccg.OrdersDetails(result)
}

//OrderHistorySingleResponse is a single order returned from a history request
type OrderHistorySingleResponse struct {
	ID              int64
	Currency        string
	Instrument      string
	OrderSide       string
	OrderType       string
	CreationTime    int64
	Status          string
	ErrorMessage    string
	Price           int64
	Volume          int64
	OpenVolume      int64
	ClientRequestID string
	Trades          []OrderHistoryTradeResponse
}

func (ohsr OrderHistorySingleResponse) convert() ccg.OrderDetails {
	return ccg.OrderDetails{
		Created:       time.Unix(ohsr.CreationTime/1000, 0),
		OrderID:       ohsr.ID,
		OrderSide:     ohsr.OrderSide,
		OrderType:     ohsr.OrderType,
		Currency:      ccg.CurrencyPair{Primary: ohsr.Instrument, Secondary: ohsr.Currency},
		VolumeOrdered: ohsr.Volume,
		VolumeFilled:  ohsr.Volume - ohsr.OpenVolume,
		Price:         ohsr.Price,
	}
}

//OrderHistoryTradeResponse is a single trade from an order in a history request
type OrderHistoryTradeResponse struct {
	ID           int64
	CreationTime int64
	Description  string
	Price        int64
	Volume       int64
	Fee          int64
}

//OrderHistory gets the users order history
func (c BTCMarketsClient) OrderHistory(Currency ccg.CurrencyPair, limit int) (ccg.OrdersDetails, error) {
	return c.OrderHistorySince(Currency, limit, 0)
}

//OrderHistorySince gets the order history since specified time (Unix time in ms)
func (c BTCMarketsClient) OrderHistorySince(Currency ccg.CurrencyPair, limit int, since int64) (ccg.OrdersDetails, error) {
	return c.orderHistory(Currency, limit, since, 1)
}

//mode;
//0 Open order history
//1 All order history
//2 Trade history
func (c BTCMarketsClient) orderHistory(Currency ccg.CurrencyPair, limit int, since int64, mode int) (ccg.OrdersDetails, error) {
	var URI string
	switch mode {
	case 0:
		URI = "/order/open"
		break
	case 1:
		URI = "/order/history"
		break
	default:
		return ccg.OrdersDetails{}, errors.New("mode somehow set incorrectly in private function")
	}
	ohr := OrderHistoryRequest{
		SecondaryCurrency: Currency.Secondary,
		PrimaryCurrency:   Currency.Primary,
		Limit:             limit,
		Since:             since,
	}
	got, err := c.signAndPost(URI, ohr)
	if err != nil {
		return ccg.OrdersDetails{}, errors.New("Error in response from server; (order history	)" + err.Error())
	}
	spew.Dump(string(got))
	var ohs OrderHistoryResponse
	err = json.Unmarshal(got, &ohs)
	if err != nil {
		return ccg.OrdersDetails{}, errors.New("Error unmarshaling response;" + err.Error() + "\n" + string(got))
	}
	if !ohs.Success {
		return ccg.OrdersDetails{}, errors.New("Error getting orders; " + ohs.ErrorMessage)
	}
	return ohs.convert(), err
}

//GetOpenOrders gets the current open orders
func (c BTCMarketsClient) GetOpenOrders(Currency ccg.CurrencyPair) (ccg.OrdersDetails, error) {
	//TODO:FIX ME
	return c.orderHistory(Currency, 200, 33434568724, 0)
	/*
		function getOpenOrders(){
			$.get("/data/trading/orders",function(data,status){
				if(data.error!=null&&data.error.length>0){
					console.log(data.error);
					}else{
						buildOrderList(data.orders);
					}
			});
		}*/
}

//OrderDetailsRequest is the struct used to request the details for order(s)
type OrderDetailsRequest struct {
	OrderIds []int `json:"orderIds"`
}

//OrdersDetailsResponse is the response recieved from order details requests
type OrdersDetailsResponse struct {
	Success      bool
	ErrorCode    int
	ErrorMessage string
	Orders       []OrderDetailsResponse
}

//OrderDetailsResponse is the details returned from a single order
type OrderDetailsResponse struct {
	ID           int
	Currency     string
	Instrument   string
	OrderSide    string
	OrderType    string
	CreationTime int64
	Status       string
	ErrorMessage string
	Price        int64
	Volume       int64
	OpenVolume   int64
	Trades       []OrderHistoryTradeResponse
}

func (odr OrderDetailsResponse) convert() ccg.OrderDetails {
	return ccg.OrderDetails{
		Created:   time.Unix(odr.CreationTime/1000, 0),
		OrderID:   int64(odr.ID),
		OrderSide: odr.OrderSide,
		OrderType: odr.OrderType,
		Currency: ccg.CurrencyPair{
			Primary:   odr.Instrument,
			Secondary: odr.Currency},
		VolumeOrdered: odr.Volume,
		VolumeFilled:  odr.Volume - odr.OpenVolume,
		Price:         odr.Price,
	}
}

//OrdersDetails gets the details of the specified orders
func (c BTCMarketsClient) OrdersDetails(orderIDs ...int) (OrdersDetailsResponse, error) {
	URI := "/order/detail"
	cor := OrderDetailsRequest{OrderIds: orderIDs}
	got, err := c.signAndPost(URI, cor)
	var odr OrdersDetailsResponse
	err = json.Unmarshal(got, &odr)
	if err != nil {
		err = errors.New("Error unmarshaling response;" + err.Error() + "\n" + string(got))
	}
	return odr, err
}

//GetOrderDetails gets a single orders details
func (c BTCMarketsClient) GetOrderDetails(orderID int) (ccg.OrderDetails, error) {
	got, err := c.OrdersDetails(orderID)
	if err != nil {
		return ccg.OrderDetails{}, err
	}
	if !got.Success {
		return ccg.OrderDetails{}, errors.New("Error getting order details; " + got.ErrorMessage)
	}
	if len(got.Orders) < 1 {
		return ccg.OrderDetails{}, errors.New("Not enough orders returned")
	}
	return got.Orders[0].convert(), nil
}

//PlaceMarketBuyOrder places a market order
// Price and volume are both *10^-8, as specified in the BTCMarkets API;
// ie: $12.34 = 1,234,000,000; 12.34BTC=1,234,000,000
func (c BTCMarketsClient) PlaceMarketBuyOrder(Currency ccg.CurrencyPair, amount int64) (ccg.PlacedOrder, error) {
	return c.createOrder(Currency.Primary, Currency.Secondary, 0, amount, true, true)
}

//PlaceLimitBuyOrder places a limit order for the specified price, that is, the price and amount will be the trades.
// Price and volume are both *10^-8, as specified in the BTCMarkets API;
// ie: $12.34 = 1,234,000,000; 12.34BTC=1,234,000,000
func (c BTCMarketsClient) PlaceLimitBuyOrder(Currency ccg.CurrencyPair, amount, price int64) (ccg.PlacedOrder, error) {
	return c.createOrder(Currency.Primary, Currency.Secondary, price, amount, true, false)
}

//PlaceMarketSellOrder places a market order
// Price and volume are both *10^-8, as specified in the BTCMarkets API;
// ie: $12.34 = 1,234,000,000; 12.34BTC=1,234,000,000
func (c BTCMarketsClient) PlaceMarketSellOrder(Currency ccg.CurrencyPair, amount int64) (ccg.PlacedOrder, error) {
	return c.createOrder(Currency.Primary, Currency.Secondary, 0, amount, false, true)
}

//PlaceLimitSellOrder places a limit order for the specified price, that is, the price and amount will be the trades.
// Price and volume are both *10^-8, as specified in the BTCMarkets API;
// ie: $12.34 = 1,234,000,000; 12.34BTC=1,234,000,000
func (c BTCMarketsClient) PlaceLimitSellOrder(Currency ccg.CurrencyPair, amount, price int64) (ccg.PlacedOrder, error) {
	return c.createOrder(Currency.Primary, Currency.Secondary, price, amount, false, false)
}
