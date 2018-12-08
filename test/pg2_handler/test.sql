CREATE USER test PASSWORD 'test';

CREATE SCHEMA test;

CREATE TABLE test.test ( id INTEGER PRIMARY KEY );

INSERT INTO test.test VALUES ( generate_series(0, 99999) );

GRANT ALL ON SCHEMA test TO test;
GRANT ALL ON test.test TO test;
