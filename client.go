package btcmarketsgo

import "encoding/base64"

//DefaultDomain is the default api domain
const DefaultDomain = "https://api.btcmarkets.net"

//DefaultPrimaryCurrencies gets the default primary currencies
var DefaultPrimaryCurrencies = []string{"BTC", "LTC", "ETH", "ETC"}

//DefaultSecondaryCurrencies gets the secondary primary currencies
var DefaultSecondaryCurrencies = []string{"AUD"}

//DefaultCurrencies is the total list of currencies
var DefaultCurrencies = append(DefaultSecondaryCurrencies, DefaultPrimaryCurrencies...)

//WithdrawFees are the fees to withdraw from the account
var WithdrawFees = []int64{0, 0, 1000000, 1000000, 0}

//DepositFees are the fees to deposit from the account
var DepositFees = []int64{0, 0, 0, 0, 0}

//TradeFees is a matrix of the fees for the default currencies
var TradeFees = [5][5]int64{
	[5]int64{0, 22000000, 22000000, 22000000, 85000000},
	[5]int64{22000000, 0, 10000000000, 10000000000, 85000000},
	[5]int64{22000000, 10000000000, 0, 10000000000, 85000000},
	[5]int64{22000000, 10000000000, 10000000000, 0, 85000000},
	[5]int64{85000000, 85000000, 85000000, 85000000, 0},
}

var defaultAddresses = []CurrencyAddress{
	CurrencyAddress{"BTC", "1NHzJiwqwkHa97dg2dfhCKuqt1HtKdhxqL"},         //BTC
	CurrencyAddress{"LTC", "LYb1Cn9wZN4tzazqkGHSMkQeUBUHkF5V34"},         //LTC
	CurrencyAddress{"ETH", "0xc0bc7c0a8642d30f263d20966f74fc9b1960924a"}, //ETH
	CurrencyAddress{"ETC", "0x6429d27c5cc2a9da40859bdef4d77df08022d032"}, //ETC
}

//BTCMarketsClient is primary struct for interacting with the API
type BTCMarketsClient struct {
	Public              string
	decodedSecret       []byte
	Domain              string
	PrimaryCurrencies   []string
	SecondaryCurrencies []string
	Addresses           []CurrencyAddress
}

//CurrencyAddress is an entry of addresses for depositing currency
type CurrencyAddress struct {
	Currency string
	Address  string
}

//NewClient gets a new BTCMarketsClient
func NewClient(public, secret, domain string, addresses []CurrencyAddress) (*BTCMarketsClient, error) {

	data, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, err
	}
	return &BTCMarketsClient{
		Public:        public,
		decodedSecret: data,
		Domain:        domain,
		Addresses:     addresses,
	}, nil
}

//NewDefaultClient gets a new client with default settings. NOTE: ADDRESSES ARE ALSO SET
func NewDefaultClient(public, secret string, err error) (*BTCMarketsClient, error) {
	if err != nil {
		return nil, err
	}
	return NewClient(public, secret, DefaultDomain, defaultAddresses)
}
