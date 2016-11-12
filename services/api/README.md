Botbox API
==========

This service is responsible for the REST API that all botbox services can
use to communicate with the database.

Usage
-----
Ensure the following environment variables are defined on your host and the
database service is running:
```
POSTGRES_DB_USER=botbox
POSTGRES_DB_NAME=botbox
POSTGRES_DB_PASSWORD=???
POSTGRES_DB_HOST=???

SMTP_IDENTITY=
SMTP_USERNAME=???
SMTP_PASSWORD=???
SMTP_HOST=???
SMTP_PORT=25

BOTBOX_DOMAIN_NAME=example.com

BOTBOX_RECAPTCHA_SECRET=???
```

```$ cd server/```

```$ ./build.sh && ./run.sh```

This will start up a docker container running a HTTP server
listening on port 8081.

Making Requests
===============

Requests to the API which require authentication must have the following
HTTP headers set:

```X-Request-ID```: The session token given by the server when signing in with
/session/auth

```Date```: The date the session was given in RFC 7231 format.

```Authentication```: The authentication header must be of the form
"HMAC256 <encrypted mac>" where <encryted mac> is created by hashing the
concatenation of: date (the same as the header) + session token + request body
using HMAC with the session token.
