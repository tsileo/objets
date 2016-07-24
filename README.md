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

## License

MIT, see LICENSE
