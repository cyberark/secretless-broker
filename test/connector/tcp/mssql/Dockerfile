FROM secretless-dev

# apt-get and system utilities
RUN apt-get update && apt-get install -y \
	curl apt-transport-https debconf-utils \
    && rm -rf /var/lib/apt/lists/*

# Add custom MS repository
RUN curl https://packages.microsoft.com/keys/microsoft.asc | apt-key add -
RUN curl --fail \
	--retry 5 \
	--retry-max-time 10 \
	https://packages.microsoft.com/config/debian/10/prod.list | tee /etc/apt/sources.list.d/mssql-release.list

# Install SQL Server drivers and tools
RUN apt-get update && ACCEPT_EULA=Y apt-get install -y libodbc1 unixodbc msodbcsql17 mssql-tools unixodbc-dev
ENV PATH $PATH:/opt/mssql-tools/bin

# Install and set locale to en_US.UTF-8
#
# sqlcmd expects the en_US.UTF-8 locale to be available otherwise it'll throw the following error:
# terminate called after throwing an instance of 'std::runtime_error'
#  what():  locale::facet::_S_create_c_locale name not valid
#

# Install locales package
RUN apt-get -y install locales
# Uncomment en_US.UTF-8 for inclusion in generation
RUN sed -i 's/^# *\(en_US.UTF-8\)/\1/' /etc/locale.gen
# Generate en_US.UTF-8 locale
RUN locale-gen en_US.UTF-8
# Set locale to en_US.UTF-8
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

# Add python 3 and pyodbc
RUN apt-get install -y python3 python3-pip
RUN pip3 install pyodbc

# Add java8 and add to $PATH
# Fix cert issues
RUN apt-get update && \
    apt-get install -y ant \
                       software-properties-common \
                       ca-certificates-java && \
    apt-add-repository 'deb http://security.debian.org/debian-security stretch/updates main' && \
    apt-get update && \
    apt-get install -y openjdk-8-jdk && \
    apt-get clean && \
    update-ca-certificates -f


# Setup JAVA_HOME -- useful for docker commandline
ENV JAVA_HOME /usr/lib/jvm/java-8-openjdk-amd64/
