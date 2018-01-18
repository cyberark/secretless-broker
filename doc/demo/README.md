# Secretless Demo

Start Conjur and load the policies:

1. `$ docker-compose build`
2. `$ docker-compose up -d pg conjur`
3. `$ docker-compose exec conjur conjurctl role retrieve-key dev:user:admin` to get the Admin API Key
4. `$ docker-compose run --rm cli5`
5. `cli5:/work# conjur authn login admin`
5. `cli5:/work# conjur policy load root policy/conjur.yml`

Bring up the `admin` container. Via secretless, use it to create an RDS database. Then store the database URL, username and password in Conjur.

1. `$ export AWS_ACCESS_KEY_ID=<your-access-key>`
2. `$  export AWS_SECRET_ACCESS_KEY=<your-secret>`
3. `$  export CONJUR_AUTHN_API_KEY=<api key for alice>`
4. `$ docker-compose up admin_secretless`
5. `$ docker-compose run --rm admin`
6. `admin:/work# password=$(openssl rand -hex 12)`
7. `admin:/work# ./create_db_instance.sh testdb $password`
8. `admin:/work# ./wait_for_db_instance.sh`
9. `admin:/work# ./store_db_password.sh $password`

Now bring up the `myapp` container with its Secretless provider. 

1. `$  export CONJUR_AUTHN_API_KEY=<api key for host/myapp>`
2. `$ docker-compose up myapp_secretless`

Connect to the database and print the data in `test.t`:

```sh-session
$ dc run --rm --no-deps myapp
root@a976f9ea6901:/# psql postgres
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
