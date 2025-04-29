# fadingMetricsCache

fadingMetricsCache is a library that provides for tracking of metrics scrapers so that events can be delivered exactly once to each scraper.

The aim of this lib is to support exactly once delivery for metrics that are normally presented as a persistent data series.

## Parameters

`scraperTimeout` is the time to continue tracking a scraper that has not been seen. When this timeout is reached, the scraper is dropped from tracking. Scrapers must be tracked in order for events to be collected for them to scrape. The first time a scraper is seen, it will receive a dedicated cache to collect events for that scraper.

[IGNORED] `maxScrapers` is the max allowed scrapers to track.

[IGNORED] `maxCacheKeys` is the max events to track for each scraper. If exceeded, events will be dropped until cleared by the next scrape.

## Example Use

```
cache := FadingMetricsCache{} // not usable yet

cache.Configure(context.Background(), time.Second*5, 2, 1000) // usable

cache.Scrape("scraper_identity") // this scraper is now registered

// any events that are registered will be stored for each tracked scraper
// note that the value is 1 for this data point... it could be any floating point number
// note timestamp is NOT used in order to induce staleness marker in Prometheus
cache.RegisterValue("events_to_metrics", map[string]string{"test":"verify_thing","result": "pass"}, 1, false)

// scrape with a new scraper, it will be tracked but first scrape will be empty
assert.Equal(t, 0, len(cache.Scrape("new")), "new scraper is now registered, no entries yet")

// scrape twice with already-present scraper, it will deliver the received event only once
assert.Equal(t, 1, len(cache.Scrape("scraper_identity")), "scraper is registered, has example event")
assert.Equal(t, 0, len(cache.Scrape("scraper_identity")), "scraper is registered, no new values since last scrap")
```

## Tests

### Configuration

Configure is required.

- TestConfigure
- TestConfigureIsRequiredForRegisterValue
- TestConfigureIsRequiredForScrape

### Value and Scraper Registration

Caches are not created to track events until scrapers are being tracked.

- TestRegisterValue
- TestRegisterBadValue
- TestScraper

### Clearing By Reading and Timeout

Reading by one scraper clears that scrapers cache. Timeout prevents dead scrapers from consuming memory to track events that won't be retrieved.

- TestScraperClear
- TestScraperTimeout
- TestMultiScraperTimeout
- TestScraperTimeoutCancellation

### Test Overwrite Vs New Entry

Multiple matching events (identical labels) seen in the same scrape interval will overwrite the previous version.

- TestUniqueKey

### Format

Verify inclusion of timestamp.

- TestScrapeFormat
