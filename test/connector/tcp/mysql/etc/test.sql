-- Create a secondary test user (previously used mysql_native_password, now uses server default)
CREATE USER 'testuser_native_password'@'localhost' IDENTIFIED BY 'testpass';
GRANT ALL PRIVILEGES ON *.* TO 'testuser_native_password'@'localhost';

CREATE USER 'testuser_native_password'@'%' IDENTIFIED BY 'testpass';
GRANT ALL PRIVILEGES ON *.* TO 'testuser_native_password'@'%';

-- Create a user with the new caching_sha2_password plugin which is the default in MySQL 8.0
CREATE USER 'testuser'@'localhost' IDENTIFIED BY 'testpass';
GRANT ALL PRIVILEGES ON *.* TO 'testuser'@'localhost';

CREATE USER 'testuser'@'%' IDENTIFIED BY 'testpass';
GRANT ALL PRIVILEGES ON *.* TO 'testuser'@'%';

CREATE DATABASE testdb;

CREATE TABLE testdb.test ( id int );

INSERT INTO testdb.test VALUES ( 1 ), ( 2 );
