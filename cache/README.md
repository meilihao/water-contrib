# cache

Middleware cache provides cache management for [water](https://github.com/meilihao/water). It can use many cache adapters, including memory, ssdb.

### Installation

	go get github.com/meilihao/water-contrib/cache

## Adapter

### memory adapter

Configure memory adapter like this:
```json
{"Interval":60}
```
interval means the gc time. The cache will check at each time interval, whether item has expired.

### ssdb adapter(recommend)

Configure ssdb adapter like this:
```json
{"Host":"127.0.0.1","Port":8888,"MinPoolSize":5,"MaxPoolSize":50,"AcquireIncrement":5,"Prefix":"cssdb_"}
```
Prefix is the prefix of ssdb key.

## Getting Help

- [API Reference](https://gowalker.org/github.com/meilihao/water-contrib/cache)

## Credits

This package is a modified version of [go-macaron/cache](https://github.com/go-macaron/cache).

## License

This project is under the Apache License, Version 2.0. See the [LICENSE](LICENSE) file for the full license text.
