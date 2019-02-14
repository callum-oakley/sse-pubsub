# sse-pubsub

A single process, in memory [pubsub][] service, using the [server-sent events
protocol][sse].

## API

### Subscribe

Subscribe to a channel by making an HTTP GET request.

```
GET /:channel
```

Publishes are streamed over the response body [as described here][sse-format].

Clients which understand SSE natively can handle the format for us. For
example, in the browser one can use the [EventSource][] interface.

```js
new EventSource(`${host}/${channel}`).onmessage = ({ data }) =>
  console.log("Received data", JSON.parse(data))
```

### Publish

Publish to a channel by making an HTTP POST request. The body of the request
will be sent to all that channel's subscribers.

```
POST /:channel '{ "some": "data" }'
```

For example, in the browser.

```js
fetch(`${host}/${channel}`, {
  method: "POST",
  body: JSON.stringify({ some: "data" }),
})
```

## Run locally

With a working [Go installation][go], set a port and use `go run`.

```
$ PORT=5000 go run main.go
2019/02/14 12:40:32 listening on :5000
```

### Test with curl

Subscribe:

```
$ curl http://localhost:5000/
```

Then publish:

```
$ curl http://localhost:5000/ -d 'Hello world!'
```

Then we should get the message on the subscription:

```
data: Hello world!
```

### Test in the browser

Subscribe:

```js
new EventSource(`http://localhost:5000/`).onmessage = ({ data }) =>
  console.log(`Received data '${data}'`)
```

Then publish:

```js
fetch("http://localhost:5000/", {
  method: "POST",
  body: "Hello world!",
})
```

Then we should get the message on the subscription:

```
Received data 'Hello world!'
```

[pubsub]: https://en.wikipedia.org/wiki/Publish%E2%80%93subscribe_pattern
[sse]: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events
[sse-format]: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#Event_stream_format
[EventSource]: https://developer.mozilla.org/en-US/docs/Web/API/EventSource
[go]: https://golang.org/doc/install
