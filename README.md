### Hexlet tests and linter status:
[![Actions Status](https://github.com/darkartx/go-project-316/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/darkartx/go-project-316/actions)
[![CI](https://github.com/darkartx/go-project-316/actions/workflows/test.yml/badge.svg)](https://github.com/darkartx/go-project-316/actions)

### Usage
Build application  
```make build```

Run  
```./bin/hexlet-go-crawler <url>```

### Help
```
NAME:
   hexlet-go-crawler - analyze a website structure

USAGE:
   hexlet-go-crawler [global options] [arguments...]

GLOBAL OPTIONS:
   --depth uint         crawl depth (default: 10)
   --retries uint       number of retries for failed requests (default: 1)
   --delay duration     delay between requests (example: 200ms, 1s) (default: 0s)
   --timeout duration   per-request timeout (default: 15s)
   --rps uint           limit requests per second (overrides delay) (default: 0)
   --user-agent string  custom user agent
   --workers uint       number of concurrent workers (default: 4)
   --help, -h           show help
```

## Depth

The `depth` parameter defines the crawling depth. The starting URL is assigned a depth of 0 in the final report, with subsequent addresses having increased depth values based on their distance from the origin.

## Retries

The `retries` parameter specifies the number of retry attempts for failed HTTP requests. If a request fails due to a network error or a server-side issue, the crawler will retry the request up to the specified number of times before marking the URL as failed. The default value is `1`, meaning each URL will be attempted once without any additional retries.
