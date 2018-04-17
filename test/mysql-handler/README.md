# MySQL Handler Development

# Quick Test (Laptop Environment)

You can use Secretless to connect a client to a MySQL database, without the client knowing the database password.

Run MySQL server in Docker using `docker-compose`:

```sh-session
$ cd doc/quick-demo/
$ docker-compose up -d mysql
Creating network "quickdemo_default" with the default driver
Creating quickdemo_mysql_1 ... done
```

Verify that MySQL is running and accepting connections on port 3306:

```
$ docker-compose ps
      Name                 Command             State              Ports       
------------------------------------------------------------------------------
quickdemo_mysql_1   /entrypoint.sh mysqld   Up (healthy)   3306/tcp, 33060/tcp
```

Now you can test a normal connection to MySQL in which the client knows the password. Start a `dev` container:

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

This is the normal way of connecting to MySQL. Now let's see how to connect a client to the database without knowing the password.
