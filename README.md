CProxy By Nathan Ogden
======================

About
-----

CProxy is an extendable web proxy that can sit between your web application and web server. It acts
as a simple passthrough that can be extended to manipulate the results of the data coming from the web
server. The intended use is for caching and optmization.


Building
--------

CProxy was written with Golang. As such make sure you have Golang 1.8+ installed before building the application.

```
go get github.com/alash3al/go-fastcgi-client
go build
```