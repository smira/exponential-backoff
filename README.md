Exponental Backoff
===================

Exponential backoff vs. simple backoff simulation in Go.

Part of highload and reliability workshop (часть мастер-класса про высокие нагрузки и надежность)
http://smira.highload.ru/ (in Russian).

Building
--------

You need go 1.1+ to build this program. Go can be downloaded from http://golang.org/doc/install or
installed as package for your OS.

To build:

    go build -o client client.go
    go build -o server server.go

You should have two programs: `client` and `server`.

Running
-------

First, it is recommended to raise limits on number of open files, e.g.:

    ulimit -n 10000

Start server:

    ./server

In different terminal, try starting client:

    ./client

This would start client with simple (fixed) delay in case of request error. Try
stopping server with `^Z`. Watch client discovering server being unavailable,
and bring server back by running command `fg`. See how server "dies" under
storm of connections from the client. Repeat the same with exponential backoff
in client:

    ./client -exponential-backoff

A lot of options, timeouts, etc. could be controlled with command-line flags:

    ./client -h
    ./server -h

Have fun!
