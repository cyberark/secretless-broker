FROM mcr.microsoft.com/mssql-tools:latest

ENV PATH $PATH:/opt/mssql-tools/bin

# Install wget.
#
# Note that the build of the base image above includes the addition of a
# custom Microsoft distro to the apt-get source list (see:
# https://github.com/Microsoft/mssql-docker/blob/master/linux/mssql-tools/Dockerfile).
# In the past, we have seen transient failures for `apt-get update` that were
# caused by transient checksum errors in this distro:
#      Reading package lists...
#      E: Failed to fetch https://packages.microsoft.com/ubuntu/16.04/
#      prod/dists/xenial/main/binary-amd64/Packages.gz  Hash Sum mismatch
# Since the install of wget does not require any updates to the packages in
# this distro, we remove this distro list before doing `apt-get update` so
# that we're agnostic to these transient checksum errors.
RUN rm -rf /etc/apt/sources.list.d/mssql-release.list && \
    apt-get update && \
    apt-get install -y wget
