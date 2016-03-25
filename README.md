# maxwells-daemon

[![Build Status](https://img.shields.io/travis/thumbtack/maxwells-daemon.svg)](https://travis-ci.org/thumbtack/maxwells-daemon)
[![BSD 3 Clause License](https://img.shields.io/badge/license-BSD--3--Clause-blue.svg)](https://github.com/thumbtack/maxwells-daemon/blob/master/LICENSE.txt)

maxwells-daemon is a tool for canarying traffic to different versions of an
application.

The daemon is designed to integrate with a proxy or web server; the server will
communicate with the daemon over TCP and use the response to determine where to
send requests.

## Building

The daemon can be built using the same process as any Go executable:

```
go get ./... && go build
```

## Running

By default, the daemon is powered by a DynamoDB backend. The flags `-region`,
`-table`, and `-application` are used to determine where to find the rollout
percentage. DynamoDB tables are expected to have a string hash key of
"application", a string range key of "version" (which must always have the
value "canary"), and a number key "rollout" that holds a value in the range
[0.0,1.0]. The credentials for AWS should be provided through the environment
variables `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`.

An example execution of the program would be:

```
AWS_ACCESS_KEY_ID=AKID1234567890 \
AWS_SECRET_ACCESS_KEY=MY-SECRET-KEY \
maxwells-daemon \
    -application 'website' \
    -table 'MaxwellsDaemon' \
    -region 'us-east-1'
```

The daemon defaults to sending statsd metrics on the default statsd port to
track performance.

If any hiccups occur while communicating with DynamoDB, the daemon will default
to a 0% rollout (this avoids the situation where a canary is unable to be
reverted).

The `examples/` directory contains a sample systemd service file for the
daemon.

## Integrations

The daemon accepts connections over the socket
`unix:/tmp/maxwells-daemon.sock`. Initial requests to the daemon should send a
single newline character. The daemon will always respond with two
`\n`-terminated lines. The first will be the assignment, which should be
provided during all subsequent requests. The second will be the location,
either "master" or "canary", specifying if the request should be canaried.

An example Nginx integration may be found in the `examples/` directory.

## Extensibility

The daemon is made of four components: rollout, server, handler, and monitor.
See the respective `.go` files for information on the interface and usage of each
component. Implementing a custom rollout backend, server, or monitoring system is as simple as
adding code that correctly implements the interface and switching the reference
in `main.go`.
