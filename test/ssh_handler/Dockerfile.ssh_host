FROM alpine:latest

RUN apk add --update openssh

COPY ./id_insecure.pub /tmp/id_insecure.pub

CMD ["/usr/sbin/sshd", "-D"]

# Root account is locked from logging in by default so
# we unlock it
RUN sed -i s/^root:!/"root:*"/g /etc/shadow

RUN ssh-keygen -A && \
    mkdir -p /root/.ssh && \
    chmod 700 /root/.ssh && \
    cat /tmp/id_insecure.pub >> /root/.ssh/authorized_keys
