GRANT ALL PRIVILEGES ON *.* TO 'testuser'@'localhost' IDENTIFIED BY 'testpass';
GRANT ALL PRIVILEGES ON *.* TO 'testuser'@'%' IDENTIFIED BY 'testpass';

CREATE DATABASE testdb;

CREATE TABLE testdb.test ( id int );

USE testdb;
DROP PROCEDURE IF EXISTS prepare_data;
DELIMITER $$
CREATE PROCEDURE prepare_data()
BEGIN
  DECLARE i INT DEFAULT 1;

  WHILE i < 100000 DO
    INSERT INTO testdb.test (id) VALUES (i);
    SET i = i + 1;
  END WHILE;
END$$
DELIMITER ;

CALL prepare_data();
