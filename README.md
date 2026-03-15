### Hexlet tests and linter status:
[![Actions Status](https://github.com/darkartx/go-project-316/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/darkartx/go-project-316/actions)

### Usage
Build application
`make build`

Run
`./bin/hexlet-go-crawler <url>

### Help
`NAME:
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
   --help, -h           show help`