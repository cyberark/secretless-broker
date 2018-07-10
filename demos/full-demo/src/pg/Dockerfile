FROM phusion/baseimage

COPY ./id_insecure.pub /tmp/id_insecure.pub
RUN cat /tmp/id_insecure.pub >> /root/.ssh/authorized_keys

RUN rm -f /etc/service/sshd/down

# Regenerate SSH host keys. baseimage-docker does not contain any, so you
# have to do that yourself. You may also comment out this instruction; the
# init system will auto-generate one during boot.
RUN /etc/my_init.d/00_regen_ssh_host_keys.sh

RUN groupadd -r postgres --gid=999 && useradd -r -g postgres --uid=999 postgres

RUN apt-get update && apt-get install -y postgresql libpq-dev python3-psycopg2

RUN mkdir /etc/service/pg
COPY ./runit/ /etc/service/pg/

COPY pg_hba.conf /etc/postgresql/9.5/main/pg_hba.conf

ENV PGDATA /var/lib/postgresql/data

RUN echo "listen_addresses = '*'" >> /etc/postgresql/9.5/main/postgresql.conf
RUN mkdir -p /var/run/postgresql && chown -R postgres:postgres /var/run/postgresql && chmod 2777 /var/run/postgresql
RUN mkdir -p "$PGDATA" && chown -R postgres:postgres "$PGDATA" && chmod 777 "$PGDATA"
RUN mkdir -p /var/run/postgresql/9.5-main.pg_stat_tmp
RUN chown postgres:postgres /var/run/postgresql/9.5-main.pg_stat_tmp -R
