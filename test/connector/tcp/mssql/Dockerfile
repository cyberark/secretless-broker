FROM secretless-dev

# apt-get and system utilities
RUN apt-get update && apt-get install -y \
	curl apt-transport-https debconf-utils \
    && rm -rf /var/lib/apt/lists/*

# adding custom MS repository
RUN curl https://packages.microsoft.com/keys/microsoft.asc | apt-key add -
RUN curl https://packages.microsoft.com/config/ubuntu/16.04/prod.list > /etc/apt/sources.list.d/mssql-release.list

# install SQL Server drivers and tools
RUN apt-get update && ACCEPT_EULA=Y apt-get install -y msodbcsql mssql-tools
ENV PATH $PATH:/opt/mssql-tools/bin

RUN apt-get -y install locales
RUN sed -i 's/^# *\(en_US.UTF-8\)/\1/' /etc/locale.gen
RUN locale-gen en_US.UTF-8
RUN update-locale LANG=en_US.UTF-8

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
# Since the installs beyond this point will likely not require any updates to the packages in
# this distro, we remove this distro list before doing `apt-get update` so
# that we're agnostic to these transient checksum errors.
RUN rm -rf /etc/apt/sources.list.d/mssql-release.list && \
    apt-get update
