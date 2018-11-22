GRANT ALL PRIVILEGES ON *.* TO 'testuser'@'localhost' IDENTIFIED BY 'testpass';
GRANT ALL PRIVILEGES ON *.* TO 'testuser'@'%' IDENTIFIED BY 'testpass';

CREATE DATABASE testdb;

CREATE TABLE testdb.test ( id int );

INSERT INTO testdb.test VALUES ( 1 ), ( 2 );
