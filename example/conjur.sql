CREATE USER conjur PASSWORD 'conjur';

CREATE SCHEMA conjur;

CREATE TABLE conjur.test ( id integer );

INSERT INTO conjur.test VALUES ( 1 ), ( 2 );

GRANT ALL ON SCHEMA conjur TO conjur;
GRANT ALL ON conjur.test TO conjur;
