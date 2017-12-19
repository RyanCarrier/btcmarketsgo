package btcmarketsgo

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

func getBody(request string) ([]byte, error) {
	log.Debug("Doing request:", "GET", ", address:", request)
	resp, err := http.Get(request)
	if err != nil {
		return []byte{}, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	log.Debug("Response status:", strconv.Itoa(resp.StatusCode), " Response:", string(body))
	return body, err
}

func (c BTCMarketsClient) sign(URI, body string) (int64, string) {
	now := time.Now().Unix() * 1000 //milliseconds
	return now, c.hashEncode(URI + "\n" + strconv.FormatInt(now, 10) + "\n" + body)
}

func (c BTCMarketsClient) hashEncode(message string) string {
	mac := hmac.New(sha512.New, c.decodedSecret)
	mac.Write([]byte(message))
	data := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return data
}

func (c BTCMarketsClient) setupHeaders(req *http.Request, timestamp int64, signature string) {
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Charset", "UTF-8")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.Public)
	req.Header.Set("timestamp", strconv.FormatInt(timestamp, 10))
	req.Header.Set("signature", signature)
}

func (c BTCMarketsClient) signAndPost(URI string, i interface{}) ([]byte, error) {
	return c.signAnd(URI, i, "POST")
}
func (c BTCMarketsClient) signAndGet(URI string) ([]byte, error) {
	return c.signAnd(URI, nil, "GET")
}

func (c BTCMarketsClient) signAnd(URI string, i interface{}, do string) ([]byte, error) {
	var body []byte
	var err error
	if i != nil {
		body, err = json.Marshal(i)
		if err != nil {
			return nil, err
		}
	} else {
		body = []byte("")
	}
	client := http.Client{}
	now, signature := c.sign(URI, string(body))
	URL := c.Domain + URI
	req, err := http.NewRequest(do, URL, bytes.NewReader(body))
	if err != nil {
		return nil, errors.New("Error creating new Request;" + err.Error())
	}
	c.setupHeaders(req, now, signature)
	log.Debug("Doing request:", string(do), ", address:", string(URL), ", body:", string(body))
	response, err := client.Do(req)
	if err != nil {
		return nil, errors.New("Error doing request;" + err.Error())
	}

	body, err = ioutil.ReadAll(response.Body)
	log.Debug("Response status:", strconv.Itoa(response.StatusCode), " Response:", string(body))
	if err != nil {
		return nil, errors.New("Error reading response;" + err.Error())
	}
	if response.StatusCode/100 != 2 {
		return nil, errors.New("StatusCode not 2xx; " + strconv.Itoa(response.StatusCode))
	}
	return body, err
}

func lookup(sl []string, cur string) int {
	for i, s := range sl {
		if s == cur {
			return i
		}
	}
	return -1
}

func lookupIndex(cur string) int {
	for i, s := range DefaultCurrencies {
		if s == cur {
			return i
		}
	}
	return -1
}
