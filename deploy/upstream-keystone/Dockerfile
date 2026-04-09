# OpenStack Keystone from Ubuntu Cloud Archive (same track as the official install guide).
# https://docs.openstack.org/keystone/latest/install/keystone-install-ubuntu.html
# Caracal UCA is supported on Jammy (22.04); Noble uses a different pocket.
# See https://wiki.ubuntu.com/OpenStack/CloudArchive
FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
	&& apt-get install -y --no-install-recommends \
		ca-certificates \
		curl \
		crudini \
		mariadb-client \
		software-properties-common \
	&& add-apt-repository -y cloud-archive:caracal \
	&& apt-get update \
	&& apt-get install -y --no-install-recommends keystone apache2 \
	&& rm -rf /var/lib/apt/lists/*

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
