# tram

Tram is a small daemon that will download and cache HTTP GET requests for later use.

# Usage

To populate content, make a POST request with a `url` field containing the url that the HTTP GET request should be made to and one or more alias fields.

    $ curl --data "alias=&url=http://ngerakines.me/" http://localhost:3000/

To fetch the content, make an HTTP GET request with the `url` query string parameter of the url that has been cached.

    $ curl http://localhost:3000/?url=http%3A%2F%2Fngerakines.me%2F

This daemon also supports HEAD requests to determine if a file has been cached or not.

    $ curl -X HEAD http://localhost:3000/?url=http%3A%2F%2Fngerakines.me%2F

When attempting GET or HEAD requests, a 404 is returned if the file has not been cached.

# License

The MIT License (MIT)

Copyright (c) 2014 Nick Gerakines
