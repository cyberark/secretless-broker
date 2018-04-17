GRANT ALL PRIVILEGES ON *.* TO 'test'@'localhost' IDENTIFIED BY 'test';
GRANT ALL PRIVILEGES ON *.* TO 'test'@'%' IDENTIFIED BY 'test';

CREATE DATABASE test;

CREATE TABLE test.test ( id int );

INSERT INTO test.test VALUES ( 1 ), ( 2 );
