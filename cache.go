package fadingMetricsCache

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const illegalChars = "\"\n"

type FadingMetricsCache struct {
	context        context.Context
	scraperTimeout time.Duration // default value is valid
	maxScrapers    int           // default value is valid
	maxCacheKeys   int           // default value is valid
	cache          map[string]map[string]string
	lastScrape     map[string]time.Time
	mu             sync.Mutex
}

func (fmc *FadingMetricsCache) Configure(context context.Context, scraperTimeout time.Duration, maxScrapers, maxCacheKeys int) {
	fmc.context = context
	fmc.scraperTimeout = scraperTimeout
	fmc.maxScrapers = maxScrapers
	fmc.maxCacheKeys = maxCacheKeys
	fmc.cache = make(map[string]map[string]string)
	fmc.lastScrape = make(map[string]time.Time)
	fmc.runScraperTimeoutGoRoutine()
}

func (fmc *FadingMetricsCache) runScraperTimeoutGoRoutine() {
	go func() { // kick off scraper timeout go routine
		if fmc.scraperTimeout == 0 {
			fmt.Println("WARN: not specifying scraper timeout is not recommended. cache can grow excessively.")
			return // no timeout
		}
		for {
			select {
			case <-fmc.context.Done():
				return
			case <-time.After(fmc.scraperTimeout):
				go func() {
					fmc.mu.Lock()
					defer fmc.mu.Unlock()
					for s := range fmc.lastScrape {
						if time.Since(fmc.lastScrape[s]) > fmc.scraperTimeout { // unregister scraper
							delete(fmc.cache, s)
							delete(fmc.lastScrape, s)
						}
					}
				}()
			}
		}
	}()
}

// registers values to already-registered scraper caches
func (fmc *FadingMetricsCache) RegisterValue(name string, labels map[string]string, value int) error {
	if fmc.cache == nil {
		panic("Configure() not called on FadingMetricsCache")
	}
	// validate label content
	for k, v := range labels {
		if tmp := strings.TrimSpace(k); tmp == "" || tmp != k {
			return fmt.Errorf("illegal leading/trailing whitespace or empty label in key: %s", k)
		}
		if tmp := strings.TrimSpace(v); tmp == "" || tmp != v {
			return fmt.Errorf("illegal leading/trailing whitespace or empty label for key %s: %s", k, v)
		}
		if strings.ContainsAny(k, illegalChars) {
			return fmt.Errorf("illegal chars in key %s", k)
		}
		if strings.ContainsAny(v, illegalChars) {
			return fmt.Errorf("illegal chars for key %s: %s", k, v)
		}
	}
	fmt.Printf("ok\n")
	// fix ordering
	keys := make([]string, len(labels))
	i := 0
	for k := range labels {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	// generate series key
	for i, k := range keys {
		keys[i] = fmt.Sprintf("%s=\"%s\"", k, labels[k])
	}
	seriesKey := fmt.Sprintf("%s{%s}", name, strings.Join(keys, ","))
	seriesValue := fmt.Sprintf("%d %d", value, time.Now().UnixMilli())

	fmc.mu.Lock()
	defer fmc.mu.Unlock()
	for scraper := range fmc.cache {
		fmc.cache[scraper][seriesKey] = seriesValue
	}
	return nil
}

// scrapes already-registered cache, else registers scraper cache
func (fmc *FadingMetricsCache) Scrape(scraperKey string) map[string]string {
	if fmc.cache == nil {
		panic("Configure() not called on FadingMetricsCache")
	}
	fmc.mu.Lock()
	defer fmc.mu.Unlock()
	for scraper := range fmc.cache {
		if scraper == scraperKey {
			// update last scrape
			fmc.lastScrape[scraperKey] = time.Now()
			// return data
			data := fmc.cache[scraperKey]
			// clear data
			fmc.cache[scraperKey] = make(map[string]string)
			// stop iterating
			return data
		}
	}
	// no scraper found, register new scraper, return empty result
	fmc.lastScrape[scraperKey] = time.Now()
	fmc.cache[scraperKey] = make(map[string]string)
	return make(map[string]string)
}
