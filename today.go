package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/wavesplatform/gowaves/pkg/client"
	"github.com/wavesplatform/gowaves/pkg/proto"
	"gopkg.in/telebot.v3"
)

type AnoteToday struct {
}

func (at *AnoteToday) sendAd(ad string) {
	var channelId int64
	if conf.Dev {
		channelId = TelDevAnoteToday
	} else {
		channelId = TelAnoteToday
	}
	r := &telebot.Chat{
		ID: channelId,
	}

	m, _ := bot.Send(r, ad, telebot.NoPreview, telebot.Silent)

	num := int64(m.ID)
	dataTransaction2("%s__adnum", nil, &num, nil)
}

func (at *AnoteToday) start() {
	for {
		if at.isNewCycle() {
			code := at.generateNewCode()

			ad := at.getAd()

			at.sendAd(fmt.Sprintf(ad, code))
		}

		time.Sleep(time.Second * MonitorTick)
	}
}

func (at *AnoteToday) isNewCycle() bool {
	ks := &KeyValue{Key: "lastAdDay"}
	db.FirstOrCreate(ks, ks)
	today := time.Now().Day()

	if ks.ValueInt != uint64(today) && time.Now().Hour() == SendAdHour {
		ks.ValueInt = uint64(today)
		db.Save(ks)

		return true
	}

	return false
}

func (at *AnoteToday) generateNewCode() int {
	ks := &KeyValue{Key: "dailyCode"}
	db.FirstOrCreate(ks, ks)

	rand.Seed(time.Now().UnixNano())
	min := 100
	max := 999

	code := rand.Intn(max-min+1) + min

	ks.ValueInt = uint64(code)
	db.Save(ks)

	return code
}

func (at *AnoteToday) getAd() string {
	ad := defaultAd

	cl, err := client.NewClient(client.Options{BaseUrl: AnoteNodeURL, Client: &http.Client{}})
	if err != nil {
		log.Println(err)
		logTelegram(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	addr, err := proto.NewAddressFromString(TodayAddress)
	if err != nil {
		log.Println(err)
		logTelegram(err.Error())
	}

	entries, _, err := cl.Addresses.AddressesData(ctx, addr)
	if err != nil {
		log.Println(err)
		logTelegram(err.Error())
	}

	var winner *string
	var amountWinner int64

	for _, e := range entries {
		if winner == nil {
			address := parseItem(e.GetKey(), 0).(string)
			amountWinner = e.ToProtobuf().GetIntValue()
			winner = &address
		} else {
			amount := e.ToProtobuf().GetIntValue()
			if amount > amountWinner {
				amountWinner = amount
				address := parseItem(e.GetKey(), 1).(string)
				winner = &address
			}
		}
	}

	adData, err := getData(AdKey, winner)
	if err != nil {
		ad = defaultAd
		log.Println(err)
		// logTelegram(err.Error())
	} else {
		adText := parseItem(adData.(string), 0)
		adLink := parseItem(adData.(string), 1)
		ad = adText.(string) + "\n\nRead <a href=\"" + adLink.(string) + "\">more</a>\n________________________\nDaily Mining Code: %d"

		winnerKey := "%s__" + *winner
		err := dataTransaction2(winnerKey, nil, nil, nil)
		if err != nil {
			log.Println(err)
			logTelegram(err.Error())
		}
	}

	return ad
}

func initAnoteToday() {
	// at := &AnoteToday{}
	// go at.start()
}

var defaultAd = `<b><u>⭕️  ANOTE 2.0 IS NOW LIVE!</u></b>    🚀

We are proud to announce that Anote 2.0 is now available for mining.

We now have our own wallet (anote.one) which is used both as a wallet and a tool for mining. Stay tuned for more exciting news, information and tutorials!

You can find tutorial how to mine here: anote.digital/mine

Join @AnoteDigital group for help and support!

________________________
Daily Mining Code: %d
`
