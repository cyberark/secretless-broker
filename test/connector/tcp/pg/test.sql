CREATE USER test PASSWORD 'test';

CREATE SCHEMA test;

CREATE TABLE test.test ( id INTEGER PRIMARY KEY );
CREATE TABLE test.encodings ( encoding TEXT, value TEXT );


INSERT INTO test.test VALUES ( generate_series(0, 99999) );
INSERT INTO test.encodings VALUES ('latin1', 'tÃ©st'); --- tést in latin1

GRANT ALL ON SCHEMA test TO test;
GRANT ALL ON test.test TO test;
GRANT ALL ON test.encodings TO test;

-- Run the following command :
-- 
-- docker-compose exec -e PGCLIENTENCODING=latin1 test psql -h secretless-dev -p 3318 -d postgres -c "select value from test.encodings where encoding='latin1'"
-- 
-- It should yield:
-- 
--  value 
-- -------
--  tést
-- (1 row)
