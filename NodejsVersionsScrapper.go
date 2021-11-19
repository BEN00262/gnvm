package main

import (
	"encoding/json"
	"log"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"go.etcd.io/bbolt"
)

func (gnvm *GNVM) GetAllNodeVersions(onFinishedCallback func([]NodeJS)) error {
	// what we do is fetch from cache the files we have
	// this should run as a server somewhere
	var found_versions []NodeJS

	/*
		{
			Async:          true,
			MaxDepth:       1,
			AllowedDomains: []string{"nodejs.org"},
		}
	*/

	collector := colly.NewCollector()

	collector.OnRequest(func(r *colly.Request) {
		log.Printf("Fetchting node versions from %s\n", NODE_VERSIONS_URL)
	})

	collector.OnError(func(r *colly.Response, e error) {
		log.Fatal("[Something went wrong :(] : ", e.Error())
	})

	// start getting the data
	collector.OnHTML("#tbVersions", func(h *colly.HTMLElement) {
		h.DOM.Children().Find("tr>td").Parent().Each(func(i int, s *goquery.Selection) {
			node_version := s.Find("td[data-label=\"Version\"]").Text()
			// npm_version := s.Find("td[data-label=\"npm\"]").Text() --> not really useful to what we are building

			// find a way to clean this up
			downloadLink := s.Find(".download-table-last").Find("a[href*=\"download/release\"]").First().AttrOr("href", "")
			// fmt.Printf("node version: %s,npm version: %s, Link: %s\n", node_version, npm_version, downloadLink)

			// save this somewhere and also list them
			split_version := strings.Split(node_version, " ")

			if len(split_version) > 2 {
				found_versions = append(found_versions, NodeJS{
					Version: "v" + split_version[len(split_version)-1],
					Link:    downloadLink,
				})
			}
		})
	})

	collector.OnScraped(func(r *colly.Response) {
		gnvm.wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()

			if err := gnvm.Db.Update(func(t *bbolt.Tx) error {
				installed_node_versions, err := t.CreateBucketIfNotExists([]byte(ALL_VERSIONS_BUCKET))

				if err != nil {
					return err
				}

				for _, nodejs := range found_versions {
					if nodejs.Version != "" {
						buf, err := json.Marshal(nodejs)
						if err != nil {
							return err
						}

						if err := installed_node_versions.Put([]byte(nodejs.Version), buf); err != nil {
							return err
						}
					}
				}

				return nil
			}); err != nil {
				log.Fatal(err.Error())
			}

		}(gnvm.wg)

		onFinishedCallback(found_versions)
	})

	return collector.Visit(NODE_VERSIONS_URL)
}
