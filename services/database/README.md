Database service
================

This service sets up and runs the database for a Botbox server, which can be
used from other containers.

Usage
-----
```$ export BOTBOX_DB_PASSWORD=<secret>```

```$ ./build.sh && ./run.sh```

This will startup an empty Postgres database listening on localhost:5432
with the table schema, etc. already configured.
