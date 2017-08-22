[![CircleCI](https://circleci.com/gh/fresh8/go-cache.svg?style=svg)](https://circleci.com/gh/fresh8/go-cache)
[![Coverage Status](https://coveralls.io/repos/github/fresh8/go-cache/badge.svg)](https://coveralls.io/github/fresh8/go-cache)
[![Go Report Card](https://goreportcard.com/badge/github.com/fresh8/go-cache)](https://goreportcard.com/report/github.com/fresh8/go-cache)

# go-cache

go-cache is a caching system for Golang with background stale cache regeneration.

go-cache is separated into:
* `cacher` - a struct that provides an entry point for getting and expiring keys for a given engine.
* `engines` - a number of different storage types, including in memory, Redis, and Aerospike.
* `joque` - a job queue using go routines and channel communication.

As go-cache is a stale cache, once an item has expired, it is not removed from the cache automatically. Instead, it will
continue to return the value currently stored, and recreate the value concurrently. Once processed, it will replace the
existing value, which will be returned by subsequent cache get requests.

An additional time value, cleanupTTL, is passed to the cacher, which is used to remove keys which have expired but not
regenerated by the given time. This stops the cache from becoming full of very old values that may not be used or, when
they are requested, return very stale data.

## Getting Started

More details are available via the godoc site:

* [cacher](https://godoc.org/github.com/fresh8/go-cache/cacher)
* engine
  * [aerospike](https://godoc.org/github.com/fresh8/go-cache/engine/aerospike)
  * [common](https://godoc.org/github.com/fresh8/go-cache/engine/common)
  * [memory](https://godoc.org/github.com/fresh8/go-cache/engine/memory)
  * [redis](https://godoc.org/github.com/fresh8/go-cache/engine/redis)
* [joque](https://godoc.org/github.com/fresh8/go-cache/joque)

### Prerequisites

* Go 1.8.x

### Installing

You can install go-cache with your favourite Go vendoring tool:

```
go get github.com/fresh8/go-cache
```

### Running

For a basic usage example, please see the [docs example folder](docs/example).

## Adding Engines

Engines follow a [clear interface](https://godoc.org/github.com/fresh8/go-cache/engine/common#Engine) exposed by
go-cache, so you can create and use your own for whichever backend you desire. Pull requests for new engines are most
definitely welcome and encouraged!

## Testing

### Prerequisites

* Glide 0.12.x

### Running Local Tests

With glide installed locally, you can use the following command to run all tests, excluding vendors:

```
go test $(glide nv)
```
