# water instrumentation example

An HTTP server using water and instrumentation. The server has a
`/users/<id>` endpoint. The server generates span information to
`stdout`.

## test
```bash
curl http://127.0.0.1:8080/users/chen
```