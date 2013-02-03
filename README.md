gandi-rrr
=========


## Overview
gandi-rrr is a web server that provides a RESTful interface for adding and removing resource records
on a domain using the [gandi.net API](http://doc.rpc.gandi.net/). Records are created by doing a
POST request to the server with a JSON payload containing a name and value. A `token` header
must also be sent with the request to identify the type of record and which domain the request is for.
Tokens are defined in the config file of the server, see included config for examples. List of supported
record types can be found [here](http://doc.rpc.gandi.net/domain/reference.html#RecordType).

## Prerequisites
None, binaries are statically linked.
If you want to compile from source you need the [go toolchain](http://golang.org/doc/install).

### Downloads
- [gandi-rrr-1.0.0-darwin-386.tar.gz](https://drive.google.com/uc?id=0B3X9GlR6EmbneXJ3OTBxSlZ3OE0)
- [gandi-rrr-1.0.0-darwin-amd64.tar.gz](https://drive.google.com/uc?id=0B3X9GlR6EmbnOTA5Z0RmRFBoVUU)
- [gandi-rrr-1.0.0-freebsd-386.tar.gz](https://drive.google.com/uc?id=0B3X9GlR6EmbnbmstNDlKRGFrVDA)
- [gandi-rrr-1.0.0-freebsd-amd64.tar.gz](https://drive.google.com/uc?id=0B3X9GlR6EmbnSTlVQnNtZnphNTQ)
- [gandi-rrr-1.0.0-linux-386.tar.gz](https://drive.google.com/uc?id=0B3X9GlR6EmbnVXItSTRfSXFWbjQ)
- [gandi-rrr-1.0.0-linux-amd64.tar.gz](https://drive.google.com/uc?id=0B3X9GlR6EmbnaWtWVm9QbFF3bUE)
- [gandi-rrr-1.0.0-linux-arm.tar.gz](https://drive.google.com/uc?id=0B3X9GlR6EmbnbVc4R016b0syUEU)
- [gandi-rrr-1.0.0-linux-arm5.tar.gz](https://drive.google.com/uc?id=0B3X9GlR6EmbneFNteXZpVkRHVGs)
- [gandi-rrr-1.0.0-windows-386.tar.gz](https://drive.google.com/uc?id=0B3X9GlR6EmbnRmpIcldEUm9BeDA)
- [gandi-rrr-1.0.0-windows-amd64.tar.gz](https://drive.google.com/uc?id=0B3X9GlR6EmbnRTZQdlJCbDN1VWM)

## Examples
All examples uses the [jq](http://stedolan.github.com/jq/) tool to pretty-print
the json response. Also note that the examples uses 'https'. gandi-rrr will only
accept https requests if `certfile` and `keyfile` are defined in the config file.
Self-signed certificates can be generated with the [cert](https://github.com/prasmussen/cert) tool.

###### List all records on domain
    $ curl --silent --insecure -H 'Token:foobarbaz' https://127.0.0.1:4430 | jq .
    {
      "records": [
        {
          "ttl": 900,
          "type": "A",
          "value": "10.0.0.100",
          "name": "foo"
        },
        {
          "ttl": 900,
          "type": "A",
          "value": "10.0.0.101",
          "name": "bar"
        }
      ],
      "domain": "bazqux.com"
    }

###### Info for specific record
    $ curl --silent --insecure -H 'Token:foobarbaz' https://127.0.0.1:4430/foo | jq .
    {
      "ttl": 900,
      "type": "A",
      "value": "10.0.0.100",
      "name": "foo"
    }

###### Add new record
    $ curl --silent --insecure -H 'Token:foobarbaz' -X POST -d '{"name": "baz", "value": "10.0.0.102"}' https://127.0.0.1:4430 | jq .
    {
      "success": true
    }

###### Delete record
    $ curl --silent --insecure -H 'Token:foobarbaz' -X DELETE https://127.0.0.1:4430/baz | jq .
    {
      "success": true
    }
