# Objets

Objets (`/ɔb.ʒɛ/`, objects in French) is an object storage server (using a directory as backend) with a AWS S3 compatible API.

## Features

 - Automatic TLS via Let's Encrypt
 - HTTP2 enabled (when using TLS)
 - support public sharing (via the `public-read` canned ACL)
 - multi-part upload support

## Constraints

 - No torrent feature
 - No ACL on bucket
 - Only support `private` and `public-read` ACL for objets

## Config

```yaml
data_dir: '/path/where/data/will/be/stored' # optional, defaults to './objets_data'

listen: ':443' # optional, defaults to ':8060', or ':433' in TLS mode

auto_tls: true # optional, defaults to 'false'. Enable/disable auto TLS via Let's Encrypt
domains: # optional. required in TLS mode. List of domains to fetch TLS certificate
  - 'objets.mydomain.com'

access_key_id: 'youracesskeyid'  # required
secret_access_key: 'yoursecretaccesskey' # required
```

## QuickStart

```sh
$ objets /path/to/config.yaml
```

## License

MIT, see LICENSE
