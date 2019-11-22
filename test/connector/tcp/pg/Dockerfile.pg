FROM postgres:9.3

COPY pg_hba.conf /etc/
COPY pg_hba test.sql /docker-entrypoint-initdb.d/
COPY ./ssl/server.pem /var/lib/postgresql/server.pem
COPY ./ssl/server-key.pem /var/lib/postgresql/server-key.pem
COPY ./ssl/ca.pem /var/lib/postgresql/ca.pem

RUN chown postgres:postgres /var/lib/postgresql/server.pem
RUN chown postgres:postgres /var/lib/postgresql/server-key.pem
RUN chown postgres:postgres /var/lib/postgresql/ca.pem
RUN chmod 0600 /var/lib/postgresql/server-key.pem
