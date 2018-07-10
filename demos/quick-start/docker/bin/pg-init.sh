#!/bin/bash
set -e

readonly QUICKSTART_DATABASE="quickstart"

psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname "${POSTGRES_DB}" <<-EOSQL
    CREATE USER ${QUICKSTART_USERNAME} WITH PASSWORD '${QUICKSTART_PASSWORD}';
    CREATE DATABASE ${QUICKSTART_DATABASE} OWNER ${QUICKSTART_USERNAME};

    \c ${QUICKSTART_DATABASE}
    set role ${QUICKSTART_USERNAME};

    CREATE TABLE counties (
      id serial PRIMARY KEY,
      name varchar (32) NOT NULL
    );

    INSERT INTO counties (name) VALUES ('Middlesex');
    INSERT INTO counties (name) VALUES ('Worcester');
    INSERT INTO counties (name) VALUES ('Essex');
    INSERT INTO counties (name) VALUES ('Suffolk');
    INSERT INTO counties (name) VALUES ('Norfolk');
    INSERT INTO counties (name) VALUES ('Bristol');
    INSERT INTO counties (name) VALUES ('Plymouth');
    INSERT INTO counties (name) VALUES ('Hampden');
    INSERT INTO counties (name) VALUES ('Barnstable');
    INSERT INTO counties (name) VALUES ('Hampshire');
    INSERT INTO counties (name) VALUES ('Berkshire');
    INSERT INTO counties (name) VALUES ('Franklin');
    INSERT INTO counties (name) VALUES ('Dukes');
    INSERT INTO counties (name) VALUES ('Nantucket');
EOSQL

touch /run/postgresql/.init