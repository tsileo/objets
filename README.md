# Objets

Objets (`/ɔb.ʒɛ/`, objects in French) is an object storage server (using a directory as back-end) with a AWS S3 compatible API.

## Features

 - Automatic TLS via Let's Encrypt
 - HTTP2 enabled (when using TLS)
 - support public sharing (via the `public-read` canned ACL)
 - multi-part upload support
 - support both AWS signature v4 and v2

### Drawbacks

 - No "one subdomain per bucket"
 - No torrent feature
 - No ACL on bucket
 - Only support `private` and `public-read` ACL for objets

## QuickStart

```sh
$ objets /path/to/config.yaml
```

### Config

```yaml
data_dir: '/path/where/data/will/be/stored' # optional, defaults to './objets_data'

listen: ':443' # optional, defaults to ':8060', or ':433' in TLS mode

tls_auto: true # optional, defaults to 'false'. Enable/disable auto TLS via Let's Encrypt
tls_domains: # optional. required in TLS mode. List of domains to fetch TLS certificate
  - 'objets.mydomain.com'

access_key_id: 'youracesskeyid'  # required
secret_access_key: 'yoursecretaccesskey' # required
```

### Make it works with s3cmd

To use **objets** with [s3cmd](http://s3tools.org/s3cmd), update `~/.s3cfg`.

#### Local server

```config
access_key = myaccesskey
secret_key = mysecretkey
host_base = localhost:8060
host_bucket = localhost:8060/%(bucket)
use_https = False
```

#### TLS mode

```config
access_key = myaccesskey
secret_key = mysecretkey
host_base = objets.yourserver.com
host_bucket = objets.yourserver.com/%(bucket)
use_https = True
```

## License

MIT, see LICENSE
