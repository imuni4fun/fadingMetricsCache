package fadingMetricsCache

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// description of test
func TestConfigure(t *testing.T) {
	cache := FadingMetricsCache{}
	cache.Configure(context.Background(), time.Second*5, 2, 1000000)
	fmt.Println("did not crash!")
}

// description of test
func TestConfigureIsRequiredForRegisterValue(t *testing.T) {
	cache := FadingMetricsCache{}
	assert.Panics(t,
		func() { _ = cache.RegisterValue("test", map[string]string{"k": "v"}, 0, true) },
		"should crash since Configure not called",
	)
}

// description of test
func TestConfigureIsRequiredForScrape(t *testing.T) {
	cache := FadingMetricsCache{}
	assert.Panics(t,
		func() { _ = cache.Scrape("a") },
		"should crash since Configure not called",
	)
}

// description of test
func TestRegisterValue(t *testing.T) {
	cache := FadingMetricsCache{}
	cache.Configure(context.Background(), time.Second*5, 2, 1000000)
	err := cache.RegisterValue("test", map[string]string{"k": "v"}, 0, true)
	assert.Nil(t, err, "RegisterValue should not return error")
	assert.Equal(t, 0, len(cache.GetScraperKeys()), "no scrapers is registered, should not be tracked")
}

// description of test
func TestRegisterBadValue(t *testing.T) {
	cache := FadingMetricsCache{}
	cache.Configure(context.Background(), time.Second*5, 2, 1000000)
	err := cache.RegisterValue("test", map[string]string{"k": "v"}, 0, true)
	assert.Nil(t, err, "RegisterValue should not return error")
	for _, c := range "\n\" " {
		only := string(c)
		before := string(c) + "valid"
		after := "valid" + string(c)
		all := []string{only, before, after}
		for _, testCase := range all {
			err := cache.RegisterValue("test", map[string]string{testCase: "v"}, 0, true)
			assert.Error(t, err, "RegisterValue should return error on invalid key: '%s'", testCase)
			err = cache.RegisterValue("test", map[string]string{"k": testCase}, 0, true)
			assert.Error(t, err, "RegisterValue should return error on invalid value: '%s'", testCase)
		}
	}
}

// description of test
func TestScraper(t *testing.T) {
	cache := FadingMetricsCache{}
	cache.Configure(context.Background(), time.Second*5, 2, 1000000)
	cache.RegisterValue("test", map[string]string{"k": "v"}, 0, true)
	assert.Equal(t, 0, len(cache.GetScraperKeys()), "no scrapers is registered, should not be tracked")
	assert.Equal(t, 0, len(cache.Scrape("a")), "no scrapers registered, no values captured") // scraper now registered
	assert.Equal(t, 1, len(cache.GetScraperKeys()), "scraper is registered, should be tracked")
	cache.RegisterValue("test", map[string]string{"k": "v"}, 0, true)
	assert.Equal(t, 1, len(cache.Scrape("a")), "scraper is registered, should capture value")
}

// description of test
func TestScraperClear(t *testing.T) {
	cache := FadingMetricsCache{}
	cache.Configure(context.Background(), time.Second*5, 2, 1000000)
	cache.RegisterValue("test", map[string]string{"k": "v"}, 0, true)
	assert.Equal(t, 0, len(cache.GetScraperKeys()), "no scrapers is registered, should not be tracked")
	assert.Equal(t, 0, len(cache.Scrape("a")), "no scrapers registered, no values captured") // scraper now registered
	cache.RegisterValue("test", map[string]string{"k": "v"}, 0, true)
	assert.Equal(t, 1, len(cache.GetScraperKeys()), "scraper is registered, should be tracked")
	assert.Equal(t, 1, len(cache.Scrape("a")), "scraper is registered, should capture value")
	assert.Equal(t, 0, len(cache.Scrape("a")), "scraper is registered, no new values")
	cache.RegisterValue("test", map[string]string{"k": "v"}, 0, true)
	assert.Equal(t, 1, len(cache.GetScraperKeys()), "scraper is registered, should be tracked")
	assert.Equal(t, 1, len(cache.Scrape("a")), "scraper is registered, should capture value")
	assert.Equal(t, 0, len(cache.Scrape("a")), "scraper is registered, no new values")
}

// description of test
func TestScraperTimeout(t *testing.T) {
	cache := FadingMetricsCache{}
	cache.Configure(context.Background(), time.Second*1, 2, 1000000)
	cache.RegisterValue("test", map[string]string{"k": "v"}, 0, true)
	assert.Equal(t, 0, len(cache.Scrape("a")), "no scrapers registered, no values captured")
	cache.RegisterValue("test", map[string]string{"k": "v"}, 0, true)
	assert.Equal(t, 1, len(cache.Scrape("a")), "scraper is registered, should capture value")
	cache.RegisterValue("test", map[string]string{"k": "v"}, 0, true)
	assert.Equal(t, 1, len(cache.GetScraperKeys()), "scraper is registered, should be tracked")
	time.Sleep(1500 * time.Millisecond) // timeout clears scrapers
	assert.Equal(t, 0, len(cache.GetScraperKeys()), "scraper should be unregistered, should not be tracked")
	assert.Equal(t, 0, len(cache.Scrape("a")), "scraper was just registered, values were cleared by timeout")
	assert.Equal(t, 1, len(cache.GetScraperKeys()), "scraper is registered, should be tracked")
}

// description of test
func TestMultiScraperTimeout(t *testing.T) {
	cache := FadingMetricsCache{}
	cache.Configure(context.Background(), time.Second*1, 2, 1000000)
	cache.RegisterValue("test", map[string]string{"k": "v"}, 0, true)
	assert.Equal(t, 0, len(cache.GetScraperKeys()), "scrapers not registered, should not be tracked")
	assert.Equal(t, 0, len(cache.Scrape("a")), "scraper not registered, no values captured")
	assert.Equal(t, 0, len(cache.Scrape("b")), "scraper not registered, no values captured")
	assert.Equal(t, 2, len(cache.GetScraperKeys()), "scrapers are registered, should be tracked")
	cache.RegisterValue("test", map[string]string{"k": "v"}, 0, true)
	assert.Equal(t, 1, len(cache.Scrape("a")), "scraper is registered, should capture value")
	assert.Equal(t, 1, len(cache.Scrape("b")), "scraper is registered, should capture value")
	cache.RegisterValue("test", map[string]string{"k": "v"}, 0, true)
	assert.Equal(t, 1, len(cache.Scrape("a")), "scraper is registered, should capture value")
	assert.Equal(t, 2, len(cache.GetScraperKeys()), "scrapers are registered, should be tracked")
	time.Sleep(1500 * time.Millisecond) // timeout clears scrapers
	assert.Equal(t, 0, len(cache.GetScraperKeys()), "scrapers not registered, should not be tracked")
	assert.Equal(t, 0, len(cache.Scrape("b")), "scraper was just registered, values were cleared by timeout")
	assert.Equal(t, 0, len(cache.Scrape("a")), "scraper was just registered, values were cleared by timeout")
	assert.Equal(t, 2, len(cache.GetScraperKeys()), "scrapers are registered, should be tracked")
}

// description of test
func TestScraperTimeoutCancellation(t *testing.T) {
	cache := FadingMetricsCache{}
	ctx, cancel := context.WithCancel(context.Background())
	cache.Configure(ctx, time.Second*1, 2, 1000000)

	// baseline test
	cache.RegisterValue("test", map[string]string{"k": "v"}, 0, true)
	assert.Equal(t, 0, len(cache.Scrape("a")), "scraper not registered, no values captured")
	cache.RegisterValue("test", map[string]string{"k": "v"}, 0, true)
	assert.Equal(t, 1, len(cache.Scrape("a")), "scraper is registered, should capture value")

	// observe cancellation
	cache.RegisterValue("test", map[string]string{"k": "v"}, 0, true)
	time.Sleep(1500 * time.Millisecond)
	assert.Equal(t, 0, len(cache.Scrape("a")), "scraper was just registered, values were cleared by timeout")

	// halt cancellation, observe durability past ususal cancellation interval
	cache.RegisterValue("test", map[string]string{"k": "v"}, 0, true)
	cancel()
	time.Sleep(1500 * time.Millisecond)
	assert.Equal(t, 1, len(cache.Scrape("a")), "scraper is still registered since cancellation check stopped, should capture value")
	assert.Equal(t, 0, len(cache.Scrape("a")), "scraper is registered, value was already scraped")
}

// description of test
func TestUniqueKey(t *testing.T) {
	cache := FadingMetricsCache{}
	cache.Configure(context.Background(), time.Second*5, 2, 1000000)
	assert.Equal(t, 0, len(cache.Scrape("a")), "no scrapers registered, no values captured") // scraper now registered
	cache.RegisterValue("test", map[string]string{"k": "v", "a": "b"}, 0, true)              // new
	cache.RegisterValue("test", map[string]string{"k": "v", "z": "b"}, 5, true)              // new
	cache.RegisterValue("test", map[string]string{"k": "v", "a": "b"}, -2, true)             // overwrite
	cache.RegisterValue("test", map[string]string{"z": "b", "k": "v"}, 1, true)              // overwrite
	data := cache.Scrape("a")
	assert.Equal(t, 2, len(data), "scraper is registered, should capture value")
	for k, v := range data {
		fmt.Printf("%s %s\n", k, v)
	}
}

// description of test
func TestScrapeFormat(t *testing.T) {
	cache := FadingMetricsCache{}
	cache.Configure(context.Background(), time.Second*5, 2, 1000000)
	assert.Equal(t, 0, len(cache.Scrape("a")), "no scrapers registered, no values captured") // scraper now registered
	cache.RegisterValue("timestamp", map[string]string{"k": "v", "a": "b"}, 0, true)
	data := cache.Scrape("a")
	assert.Equal(t, 1, len(data), "scraper is registered, should capture value")
	for k, v := range data {
		assert.Equal(t, 2, len(strings.Split(v, " ")), "value should have 2 parts: value and timestamp")
		fmt.Printf("%s %s\n", k, v)
	}
	cache.RegisterValue("no_timestamp", map[string]string{"k": "v", "a": "b"}, 0, false)
	data = cache.Scrape("a")
	assert.Equal(t, 1, len(data), "scraper is registered, should capture value")
	for k, v := range data {
		assert.Equal(t, 1, len(strings.Split(v, " ")), "value should have 1 part: value")
		fmt.Printf("%s %s\n", k, v)
	}
}
