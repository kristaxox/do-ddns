package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/digitalocean/godo"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	doToken = kingpin.Flag("do-token", "DigitalOcean token").Envar("DO_TOKEN").Required().String()
	records = kingpin.Flag("records", "A name record(s) to update").Envar("RECORDS").Required().Strings()
	domain  = kingpin.Flag("domain", "domain containing the records").Envar("DOMAIN").Required().String()
	freq    = kingpin.Flag("frequency", "Frequency to check").Envar("FREQUENCY").Default("10m").Duration()
)

func getIP() (string, error) {
	resp, err := http.Get("http://ifconfig.co")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return strings.TrimSuffix(string(body), "\n"), err
}

func main() {
	kingpin.Parse()
	client := godo.NewFromToken(*doToken)
	wg := &sync.WaitGroup{}

	for _, record := range *records {
		wg.Add(1)
		go func(client *godo.Client, record string) {
			ticker := time.NewTicker(*freq)
			for range ticker.C {
				// get current IP
				ip, err := getIP()
				if err != nil {
					logrus.WithError(err).Error("unable to get external IP")
				}
				logrus.WithFields(logrus.Fields{
					"ip": ip,
				}).Info("got current external ip")

				// get current records
				domainRecords, _, err := client.Domains.RecordsByTypeAndName(context.Background(), *domain, "A", record, nil)
				if err != nil {
					logrus.WithError(err).Error("unable to retrieve current dns record")
				}

				// iterate through records that match
				for _, domainRecord := range domainRecords {
					logrus.WithFields(logrus.Fields{
						"type":   domainRecord.Type,
						"name":   domainRecord.Name,
						"data":   domainRecord.Data,
						"domain": *domain,
					}).Info("found")

					if domainRecord.Data != ip {
						logrus.WithFields(logrus.Fields{
							"expected": ip,
							"actual":   domainRecord.Data,
							"domain":   *domain,
						}).Info("need to update IP")
						// update record with new address
						req := &godo.DomainRecordEditRequest{
							Type: "A",
							Name: domainRecord.Name,
							Data: ip,
						}
						_, _, err = client.Domains.EditRecord(context.Background(), *domain, domainRecord.ID, req)
						if err != nil {
							logrus.WithError(err).Info("unable to update record")
						} else {
							logrus.WithFields(logrus.Fields{
								"old":    domainRecord.Data,
								"new":    ip,
								"record": domainRecord.Name,
								"domain": *domain,
							}).Info("successfully updated record")
						}
					}
				}
			}
		}(client, record)
	}
	wg.Wait()
}
