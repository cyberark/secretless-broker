FROM phusion/baseimage

COPY ./id_insecure.pub /tmp/id_insecure.pub
RUN cat /tmp/id_insecure.pub >> /root/.ssh/authorized_keys

RUN rm -f /etc/service/sshd/down

ENV PGDATA /var/lib/postgresql/data

RUN ssh-keygen -A && \
    groupadd -r postgres --gid=999 && \
    useradd -r -g postgres --uid=999 postgres && \
    apt-get update && \
    apt-get install -y libpq-dev \
                       python3-psycopg2 \
                       postgresql && \
    mkdir /etc/service/pg

COPY ./runit/ /etc/service/pg/
COPY pg_hba.conf /etc/postgresql/9.5/main/pg_hba.conf

RUN echo "listen_addresses = '*'" >> /etc/postgresql/9.5/main/postgresql.conf && \
    mkdir -p /var/run/postgresql && \
    chown -R postgres:postgres /var/run/postgresql && \
    chmod 2777 /var/run/postgresql && \
    mkdir -p "$PGDATA" && \
    chown -R postgres:postgres "$PGDATA" && \
    chmod 777 "$PGDATA" && \
    mkdir -p /var/run/postgresql/9.5-main.pg_stat_tmp && \
    chown -R postgres:postgres /var/run/postgresql/9.5-main.pg_stat_tmp
