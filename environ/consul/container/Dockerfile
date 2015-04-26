# Public Domain (-) 2014-2015 The Ampify Authors.
# See the Ampify UNLICENSE file for details.

FROM espians/debian@sha256:101dbc6aeb06ae830efcc8e179ddd0b1a73730373a40b528e308d23c59a58784

# Base Packages
RUN echo "image base: 2015-04-22" && apt-get update && apt-get -y upgrade
RUN apt-get install -y ca-certificates curl python unzip

# Checksum Verifier
ADD verify-checksum /usr/local/bin/verify-checksum

# Non-root User
RUN useradd -ms /bin/bash consul
USER consul
WORKDIR /home/consul
ENV GOMAXPROCS=4 PATH=/home/consul/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

# Consul
RUN curl -L -O https://dl.bintray.com/mitchellh/consul/0.5.0_linux_amd64.zip && \
  verify-checksum 0.5.0_linux_amd64.zip 581decd401b218c181b06a176c61cb35e6e4a6d91adf3837c5d2498c7aef98d6d4da536407c800e0d0d027914a174cdb04994e5bd5fdda7ee276b168fb4a5f8e && \
  unzip 0.5.0_linux_amd64.zip && \
  mkdir bin && \
  mv consul bin/ && \
  rm 0.5.0_linux_amd64.zip

# Config
ADD consul.json /home/consul/consul.json
EXPOSE 9090 9091

# Executable
ENTRYPOINT ["/home/consul/bin/consul", "agent", "-config-file=/home/consul/consul.json"]
