package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Prices struct {
	BTC float64 `json:"BTC"`
	ETH float64 `json:"ETH"`
	USD float64 `json:"USD"`
	EUR float64 `json:"EUR"`
}

type PriceClient struct {
	Prices     *Prices
	Loaded     bool
	AnotePrice float64
}

func (pc *PriceClient) doRequest() (*Prices, error) {
	p := &Prices{}
	cl := http.Client{}

	var req *http.Request
	var err error

	req, err = http.NewRequest(http.MethodGet, PricesURL, nil)

	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}

	res, err := cl.Do(req)

	if err == nil {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		if res.StatusCode != 200 {
			log.Println(string(body))
			err := errors.New(res.Status)
			return nil, err
		}
		json.Unmarshal(body, p)
	} else {
		return nil, err
	}

	pc.Prices = p

	// pc.doRequestOrderbook()
	pc.loadPrice()

	return p, nil
}

func (pc *PriceClient) loadPrice() {
	// pc.AnotePrice = getPriceCoinGecko()
	pc.AnotePrice = 152.76
}

func (pc *PriceClient) doRequestOrderbook() {
	or := &OrderbookStatusResponse{}
	cl := http.Client{}

	var req *http.Request
	var err error

	req, err = http.NewRequest(http.MethodGet, OrderbookStatusURL, nil)

	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		log.Println(err)
		logTelegram(err.Error())
		return
	}

	res, err := cl.Do(req)

	if err == nil {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println(err)
			logTelegram(err.Error())
			return
		}
		if res.StatusCode != 200 {
			err := errors.New(res.Status)
			log.Println(err)
			logTelegram(err.Error())
			return
		}
		json.Unmarshal(body, or)
	} else {
		log.Println(err)
		logTelegram(err.Error())
		return
	}

	pc.AnotePrice = float64(or.LastPrice) / MULTI8 * pc.Prices.USD
}

func (pc *PriceClient) start() {
	for {
		if p, err := pc.doRequest(); err != nil {
			log.Println(err)
			logTelegram(err.Error())
		} else {
			pc.Prices = p
		}

		// if conf.Debug {
		// 	log.Printf("%#v\n", pc.Prices)
		// }

		pc.Loaded = true

		time.Sleep(time.Minute * 5)
	}
}

func initPriceClient() *PriceClient {
	pc := &PriceClient{Loaded: false}
	go pc.start()
	return pc
}

type OrderbookResponse struct {
	Timestamp int64 `json:"timestamp"`
	Pair      struct {
		AmountAsset string `json:"amountAsset"`
		PriceAsset  string `json:"priceAsset"`
	} `json:"pair"`
	Bids []struct {
		Amount int64 `json:"amount"`
		Price  int   `json:"price"`
	} `json:"bids"`
	Asks []struct {
		Amount int64 `json:"amount"`
		Price  int   `json:"price"`
	} `json:"asks"`
}

type OrderbookStatusResponse struct {
	Success    bool   `json:"success"`
	Ask        int    `json:"ask"`
	BidAmount  int64  `json:"bidAmount"`
	Bid        int    `json:"bid"`
	LastAmount int    `json:"lastAmount"`
	AskAmount  int64  `json:"askAmount"`
	LastSide   string `json:"lastSide"`
	Status     string `json:"status"`
	LastPrice  int    `json:"lastPrice"`
}

type AggregtorResponse struct {
	Routes []struct {
		RealPrice float64 `json:"realPrice"`
		In        int     `json:"in"`
		AllIn     int     `json:"allIn"`
		Exchanges []struct {
			From      string `json:"from"`
			To        string `json:"to"`
			Pool      string `json:"pool"`
			Type      string `json:"type"`
			AmountIn  int    `json:"amountIn"`
			AmountOut int    `json:"amountOut"`
		} `json:"exchanges"`
	} `json:"routes"`
	AggregatedProfit int     `json:"aggregatedProfit"`
	EstimatedOut     int     `json:"estimatedOut"`
	PriceImpact      float64 `json:"priceImpact"`
	Parameters       string  `json:"parameters"`
	Error            string  `json:"error"`
}

func getPriceAggregator() float64 {
	price := float64(0)
	// ar := &AggregtorResponse{}
	// cl := http.Client{}

	// var req *http.Request
	// var err error

	// req, err = http.NewRequest(http.MethodGet, AggregatorURL, nil)

	// req.Header.Set("Content-Type", "application/json")

	// if err != nil {
	// 	log.Println(err)
	// 	logTelegram(err.Error())
	// 	return price
	// }

	// res, err := cl.Do(req)

	// if err == nil {
	// 	log.Println(prettyPrint(res.Body))
	// 	body, err := ioutil.ReadAll(res.Body)
	// 	if err != nil {
	// 		log.Println(err)
	// 		logTelegram(err.Error())
	// 		return price
	// 	}
	// 	if res.StatusCode != 200 {
	// 		err := errors.New(res.Status)
	// 		log.Println(err)
	// 		logTelegram(err.Error())
	// 		return price
	// 	}
	// 	json.Unmarshal(body, ar)
	// } else {
	// 	log.Println(err)
	// 	logTelegram(err.Error())
	// 	return price
	// }

	// price = float64(ar.EstimatedOut) / 100

	price = 1.03

	return price
}

type CoinGeckoResponse struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Address string `json:"address"`
			Name    string `json:"name"`
			Symbol  string `json:"symbol"`
			// CoingeckoCoinID   any    `json:"coingecko_coin_id"`
			Decimals          int    `json:"decimals"`
			TotalSupply       string `json:"total_supply"`
			PriceUsd          string `json:"price_usd"`
			FdvUsd            string `json:"fdv_usd"`
			TotalReserveInUsd string `json:"total_reserve_in_usd"`
			VolumeUsd         struct {
				H24 string `json:"h24"`
			} `json:"volume_usd"`
			// MarketCapUsd any `json:"market_cap_usd"`
		} `json:"attributes"`
		Relationships struct {
			TopPools struct {
				Data []struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
			} `json:"top_pools"`
		} `json:"relationships"`
	} `json:"data"`
}

func getPriceCoinGecko() float64 {
	price := float64(50)

	cgr := &CoinGeckoResponse{}
	cl := http.Client{}

	var req *http.Request
	var err error

	req, err = http.NewRequest(http.MethodGet, CoinGeckoURL, nil)

	req.Header.Set("Content-Type", "application/json;version=20230302")
	req.Header.Set("Accept", "application/json;version=20230302")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

	log.Println(prettyPrint(req))

	if err != nil {
		log.Println(err)
		logTelegram(err.Error())
		return price
	}

	res, err := cl.Do(req)

	log.Println(prettyPrint(res))

	if err == nil {
		log.Println(prettyPrint(res.Body))
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println(err)
			logTelegram(err.Error())
			return price
		}
		if res.StatusCode != 200 && res.StatusCode != 304 {
			err := errors.New(res.Status)
			log.Println(err)
			log.Println(res.Body)
			logTelegram(err.Error())
			return price
		}
		json.Unmarshal(body, cgr)
	} else {
		log.Println(err)
		logTelegram(err.Error())
		return price
	}

	fdv, err := strconv.ParseFloat(cgr.Data.Attributes.FdvUsd, 32)
	if err != nil {
		log.Println(err)
		logTelegram(err.Error())
		return price
	}

	ts, err := strconv.ParseFloat(cgr.Data.Attributes.TotalSupply, 32)
	if err != nil {
		log.Println(err)
		logTelegram(err.Error())
		return price
	}

	prc, err := strconv.ParseFloat(cgr.Data.Attributes.PriceUsd, 32)
	if err != nil {
		log.Println(err)
		logTelegram(err.Error())
		return price
	}

	ts = ts / float64(MULTI8)

	price = fdv / ts

	// price = 50.47
	// prc := price

	priceInt := int64(prc * MULTI8)

	savedPriceStr, err := getData2("%s__priceAnote", nil)
	if err != nil {
		log.Println(err)
		logTelegram(err.Error())
		return price
	}

	savedPrice := savedPriceStr.(int64)

	if savedPrice != priceInt {
		dataTransaction2("%s__priceAnote", nil, &priceInt, nil)
	}

	return price
}

type DexToolsResponse struct {
	StatusCode int `json:"statusCode"`
	Data       struct {
		Price      float64 `json:"price"`
		PriceChain float64 `json:"priceChain"`
	} `json:"data"`
}

func getPriceDexTools() float64 {
	price := float64(52.92)

	// dtr := &DexToolsResponse{}
	// cl := http.Client{}

	// var req *http.Request
	// var err error

	// req, err = http.NewRequest(http.MethodGet, DexToolsURL, nil)
	// if err != nil {
	// 	log.Println(err)
	// 	logTelegram(err.Error())
	// 	return price
	// }

	// req.Header.Set("Content-Type", "application/json;version=20230302")
	// req.Header.Set("Accept", "application/json;version=20230302")
	// req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
	// req.Header.Set("X-BLOBR-KEY", "3Sd5U5Jp1YujRVRJHL1NNKiNgIOpx1PT")

	// res, err := cl.Do(req)

	// log.Println(prettyPrint(res))

	// if err == nil {
	// 	log.Println(prettyPrint(res.Body))
	// 	body, err := ioutil.ReadAll(res.Body)
	// 	if err != nil {
	// 		log.Println(err)
	// 		logTelegram(err.Error())
	// 		return price
	// 	}
	// 	if res.StatusCode != 200 && res.StatusCode != 304 {
	// 		err := errors.New(res.Status)
	// 		log.Println(err)
	// 		log.Println(res.Body)
	// 		logTelegram(err.Error())
	// 		return price
	// 	}
	// 	json.Unmarshal(body, dtr)
	// } else {
	// 	log.Println(err)
	// 	logTelegram(err.Error())
	// 	return price
	// }

	// price = dtr.Data.Price
	prc := price

	priceInt := int64(prc * MULTI8)

	savedPriceStr, err := getData2("%s__priceAnote", nil)
	if err != nil {
		log.Println(err)
		logTelegram(err.Error())
		return price
	}

	savedPrice := savedPriceStr.(int64)

	if savedPrice != priceInt {
		dataTransaction2("%s__priceAnote", nil, &priceInt, nil)
	}

	return price
}
