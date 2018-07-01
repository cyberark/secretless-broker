CREATE USER test PASSWORD 'test';

CREATE SCHEMA test;

CREATE TABLE test.test ( id integer );

INSERT INTO test.test VALUES ( 1 ), ( 2 );

GRANT ALL ON SCHEMA test TO test;
GRANT ALL ON test.test TO test;
