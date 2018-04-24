# MySQL Handler Development

## Quick Test (Laptop Environment)

Run MySQL in Docker:
```sh-session
$ docker-compose up -d mysql
```

Run Secretless locally and execute tests:
```sh-session
$ ./run_dev_test
...
ok      github.com/conjurinc/secretless/test/mysql_handler   0.048s
2018/01/11 15:06:56 Caught signal terminated: shutting down.
```

## Local Environment (Laptop)

These instructions show how to develop the Secretless MySQL handler on your local machine.

First you'll need a MySQL server. You can run one natively, or using Docker:

```sh-session
$ docker-compose up -d mysql
```

Now you can run `secretless` in a terminal:

```sh-session
$ ./run_dev
...
2018/01/10 16:33:09 mysql listener 'mysql_tcp' listening at: [::]:13306
2018/01/10 16:33:09 mysql listener 'mysql_socket' listening at: ./run/mysql/.s.MYSQL.3306
```

Now run a client in another terminal.

Connect over a Unix socket:

## Connecting to MySQL Without Secretless

You can test a normal connection to MySQL in which the client knows the password. Start a `dev` container:

```sh-session
$ docker-compose run --rm dev
Starting quickdemo_mysql_1 ... done
root@7be0ff91e64e:/#
```

Now connect to MySQL using the username "test" and password "test" (type `\q` to quit):

```sh-session
root@7be0ff91e64e:/# mysql -utest -ptest -hmysql -P3306
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 4
Server version: 5.7.21 MySQL Community Server (GPL)

Copyright (c) 2000, 2018, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> \q
```

This is the normal way of connecting to MySQL. 
