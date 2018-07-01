# Secretless RDS Demo

Start Conjur and load the policies:

1. `$ docker-compose build`
2. `$ docker-compose up -d conjur`
3. `$ docker-compose exec conjur conjurctl role retrieve-key dev:user:admin` to get the API Key for user "admin".
3. `$ docker-compose exec conjur conjurctl role retrieve-key dev:user:alice` to get the API Key for user "alice".
3. `$ docker-compose exec conjur conjurctl role retrieve-key dev:host:myapp` to get the API Key for host "myapp".

Bring up the `admin` container. Via secretless, use it to create an RDS database. Then store the database URL, username and password in Conjur.

1. `$ export AWS_ACCESS_KEY_ID=<your-access-key>`
2. `$  export AWS_SECRET_ACCESS_KEY=<your-secret>`
3. `$  export CONJUR_AUTHN_API_KEY=<api key for alice>`
4. `$ docker-compose up -d admin_secretless`
5. `$ docker-compose run --no-deps --rm admin`
6. `admin:/work# password=$(openssl rand -hex 12)`
7. `admin:/work# ./create_db_instance testdb $password`
8. `admin:/work# ./wait_for_db_instance`
9. `admin:/work# ./store_db_password $password`

Now bring up "myapp" Secretless provider. 

1. `$  export CONJUR_AUTHN_API_KEY=<api key for host/myapp>`
2. `$ docker-compose up myapp_secretless`

And the "myapp" application:

```
$ docker-compose run --rm --no-deps myapp
```

Connect to the database, load the table `test.t`, then use `psql` to interact with the database engine:

```sh-session
root@a976f9ea6901:/# cd work
root@a976f9ea6901:/work/# cat data.sql | psql postgres
CREATE SCHEMA
CREATE TABLE
INSERT 0 2
root@a976f9ea6901:/work/# psql postgres
psql (9.5.9, server 9.6.3)

postgres=> \dn;
List of schemas
  Name  | Owner
--------+-------
 public | alice
 test   | alice
(2 rows)

postgres=> select * from test.t ;
 id
----
  1
  2
(2 rows)
```

Now you can delete the database:

1. `admin:/work# ./delete_db_instance`
