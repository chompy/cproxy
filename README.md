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

Configuration
-------------

CProxy and its extensions are configurable via a JSON file, cproxy.json by default.

**proxy_type**
```
"proxy_type": "(http|fcgi)"
```
Set proxy type, either HTTP or FastCGI(fcgi).

**listen**
```
"listen": "(:<port>|<socket>)"
```
Set port or socket to listen on.

**connect**
```
"connect": "(<ip address>|<socket>)"
```
Address or socket of backend application.

**extensions.path**
```
"extensions": {
    "path" : "<path>"
}
```
Path to extensions directory, 'ext' by default.

**extensions.enabled**
```
"extensions": {
    "enabled" : ["<filename>"]
}
```
List of extensions to enable. Order matters and determines the order
of event propigation.

**extensions.config**
```
"extensions": {
    "config" : {
        "<filename>": {
            ...
        }
    }
}
```
List of configuration for each extension. See the extension README for details on
how to configure each extension.


Extensions
----------

TODO