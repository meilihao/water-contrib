# csrf 

Middleware csrf generates and validates CSRF tokens for [water](https://github.com/meilihao/water).it needs backend storage(ssdb).

## Installation

	go get github.com/meilihao/water-contrib/csrf

### Warning

1. **it depends on [water-contrib/session](github.com/meilihao/water-contrib/session)**
1. Using "Form",csrf doesn't support ReGenerateToken,`water.Context.Environ.Set` will panic,but "Header" can.

	
## Getting Help

- [API Reference](https://gowalker.org/github.com/meilihao/water-contrib/csrf)

## Example

See test case.

## License

This project is under the Apache License, Version 2.0. See the [LICENSE](LICENSE) file for the full license text.