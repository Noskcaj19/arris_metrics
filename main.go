package main

import (
	"github.com/PuerkitoBio/goquery"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	influxAddr, ok := os.LookupEnv("INFLUX_ADDR")
	if !ok {
		log.Fatalln("INFLUX_ADDR not set")
	}
	influxToken, ok := os.LookupEnv("INFLUX_TOKEN")
	if !ok {
		log.Fatalln("INFLUX_TOKEN not set")
	}
	influxOrg, ok := os.LookupEnv("INFLUX_ORG")
	if !ok {
		log.Fatalln("INFLUX_ORG not set")
	}
	influxBucket, ok := os.LookupEnv("INFLUX_BUCKET")
	if !ok {
		log.Fatalln("INFLUX_BUCKET not set")
	}
	modemAddr, ok := os.LookupEnv("MODEM_ADDR")
	if !ok {
		log.Fatalln("MODEM_ADDR not set")
	}
	rateRaw, ok := os.LookupEnv("SCRAPE_RATE_SECS")
	if !ok {
		rateRaw = "10"
	}

	rate, err := strconv.ParseInt(rateRaw, 10, 64)
	if err != nil {
		log.Panicln(err)
	}

	influxClient := influxdb2.NewClient(influxAddr, influxToken)
	writeAPI := influxClient.WriteAPI(influxOrg, influxBucket)
	errs := writeAPI.Errors()

	go func() {
		for err := range errs {
			log.Println(err)
		}
	}()

	for range time.NewTicker(time.Duration(rate) * time.Second).C {
		err := reportChannels(modemAddr, writeAPI)
		if err != nil {
			log.Println(err)
		}
	}
}

func reportChannels(modemAddr string, writeAPI api.WriteAPI) error {
	resp, err := http.Get(modemAddr + "/cgi-bin/status")
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	downstreamChannels := extractDownstreamChannels(doc)
	reportDownstreamEntries(downstreamChannels, writeAPI)

	upstreamChannels := extractUpstreamChannels(doc)
	reportUpstreamEntries(upstreamChannels, writeAPI)

	return nil
}

func reportDownstreamEntries(entries []map[string]string, writeAPI api.WriteAPI) {
	for _, entry := range entries {
		tags := make(map[string]string)
		fields := make(map[string]interface{})

		if channelId, ok := entry["Channel ID"]; ok {
			tags["channel_id"] = channelId
		} else {
			log.Panicln("No channel id")
		}

		if channelRaw, ok := entry["Channel"]; ok {
			if channel, err := strconv.Atoi(channelRaw); err == nil {
				fields["channel"] = channel
			} else {
				log.Panicln(err)
			}
		}

		if lockStatus, ok := entry["Lock Status"]; ok {
			fields["lock_status"] = lockStatus
		}

		if modulation, ok := entry["Modulation"]; ok {
			fields["modulation"] = modulation
		}

		if rawFreq, ok := entry["Frequency"]; ok {
			if freq, err := parseFreq(rawFreq); err == nil {
				fields["frequency"] = freq
			} else {
				log.Panicln(err)
			}
		}

		if rawPower, ok := entry["Power"]; ok {
			if power, err := parsePowerLevel(rawPower); err == nil {
				fields["power"] = power
			} else {
				if _, ok := err.(noneError); !ok {
					log.Panicln(err)
				}
			}
		}

		if rawSNR, ok := entry["SNR"]; ok {
			if snr, err := parseSNR(rawSNR); err == nil {
				fields["snr"] = snr
			} else {
				if _, ok := err.(noneError); !ok {
					log.Panicln(err)
				}
			}
		}

		if correctedRaw, ok := entry["Corrected"]; ok {
			if corrected, err := strconv.Atoi(correctedRaw); err == nil {
				fields["corrected"] = corrected
			} else {
				log.Panicln(err)
			}
		}

		if uncorrectableRaw, ok := entry["Uncorrectables"]; ok {
			if uncorrectable, err := strconv.Atoi(uncorrectableRaw); err == nil {
				fields["uncorrectable"] = uncorrectable
			} else {
				log.Panicln(err)
			}
		}

		p := influxdb2.NewPoint("downstream_channels",
			tags,
			fields,
			time.Now())
		writeAPI.WritePoint(p)
	}
}

func extractDownstreamChannels(doc *goquery.Document) []map[string]string {
	rows := doc.Find("#bg3 > div.container > div.content > form > center:nth-child(5) > table > tbody > tr:nth-child(n+3)")
	headers := make([]string, 0)
	doc.Find("#bg3 > div.container > div.content > form > center:nth-child(5) > table > tbody > tr:nth-child(2)").Children().Each(func(i int, s *goquery.Selection) {
		headers = append(headers, strings.TrimSpace(s.Text()))
	})

	out := make([]map[string]string, 0)

	rows.Each(func(ri int, row *goquery.Selection) {
		rowData := make(map[string]string)
		row.Children().Each(func(ci int, entry *goquery.Selection) {
			rowData[headers[ci]] = strings.TrimSpace(entry.Text())
		})
		out = append(out, rowData)
	})

	return out
}

func reportUpstreamEntries(entries []map[string]string, writeAPI api.WriteAPI) {
	for _, entry := range entries {
		tags := make(map[string]string)
		fields := make(map[string]interface{})

		if channelId, ok := entry["Channel ID"]; ok {
			tags["channel_id"] = channelId
		} else {
			log.Panicln("No channel id")
		}

		if channelRaw, ok := entry["Channel"]; ok {
			if channel, err := strconv.Atoi(channelRaw); err == nil {
				fields["channel"] = channel
			} else {
				log.Panicln(err)
			}
		}

		if lockStatus, ok := entry["Lock Status"]; ok {
			fields["lock_status"] = lockStatus
		}

		if lockStatus, ok := entry["US Channel Type"]; ok {
			fields["channel_type"] = lockStatus
		}

		if rawSymbolRate, ok := entry["Symbol Rate"]; ok {
			if symbolRate, err := parseSymbolRate(rawSymbolRate); err == nil {
				fields["symbol_rate"] = symbolRate
			} else {
				log.Panicln(err)
			}
		}

		if rawFreq, ok := entry["Frequency"]; ok {
			if rawFreq == "----" {
				log.Panicf("%#v\n", entry)
			}
			if freq, err := parseFreq(rawFreq); err == nil {
				fields["frequency"] = freq
			} else {
				log.Panicln(err)
			}
		}

		if rawPower, ok := entry["Power"]; ok {
			if power, err := parsePowerLevel(rawPower); err == nil {
				fields["power"] = power
			} else {
				if _, ok := err.(noneError); !ok {
					log.Panicln(err)
				}
			}
		}

		p := influxdb2.NewPoint("upstream_channels",
			tags,
			fields,
			time.Now())
		writeAPI.WritePoint(p)
	}
}

func extractUpstreamChannels(doc *goquery.Document) []map[string]string {
	rows := doc.Find("#bg3 > div.container > div.content > form > center:nth-child(8) > table >  tbody > tr:nth-child(n+3)")
	headers := make([]string, 0)
	doc.Find("#bg3 > div.container > div.content > form > center:nth-child(8) > table > tbody > tr:nth-child(2)").Children().Each(func(i int, s *goquery.Selection) {
		headers = append(headers, strings.TrimSpace(s.Text()))
	})

	out := make([]map[string]string, 0)

	rows.Each(func(ri int, row *goquery.Selection) {
		rowData := make(map[string]string)
		row.Children().Each(func(ci int, entry *goquery.Selection) {
			rowData[headers[ci]] = strings.TrimSpace(entry.Text())
		})
		out = append(out, rowData)
	})

	return out
}
