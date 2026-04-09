#!/bin/bash
set -euo pipefail

DB_HOST="${KEYSTONE_DB_HOST:-mariadb_upstream}"
DB_USER="${KEYSTONE_DB_USER:-keystone}"
DB_PASS="${KEYSTONE_DB_PASS:-keystonepass}"
DB_NAME="${KEYSTONE_DB_NAME:-keystone}"
ADMIN_PASS="${KEYSTONE_ADMIN_PASSWORD:-admin}"
PUBLIC_URL="${KEYSTONE_PUBLIC_URL:-http://127.0.0.1:15000/v3/}"

CONN="mysql+pymysql://${DB_USER}:${DB_PASS}@${DB_HOST}/${DB_NAME}"

echo "Waiting for MariaDB at ${DB_HOST}..."
for _ in $(seq 1 90); do
	if mysql -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASS" -e "SELECT 1" >/dev/null 2>&1; then
		break
	fi
	sleep 2
done
if ! mysql -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASS" -e "SELECT 1" >/dev/null 2>&1; then
	echo "MariaDB not reachable" >&2
	exit 1
fi

if ! grep -q '^ServerName' /etc/apache2/apache2.conf 2>/dev/null; then
	echo 'ServerName localhost' >>/etc/apache2/apache2.conf
fi

crudini --set /etc/keystone/keystone.conf database connection "$CONN"
crudini --set /etc/keystone/keystone.conf token provider fernet
# Avoid PermissionError when keystone-manage runs as user keystone (package defaults log to /var/log/keystone).
crudini --set /etc/keystone/keystone.conf DEFAULT log_file ''
crudini --set /etc/keystone/keystone.conf DEFAULT use_stderr 'true'

install -d -o keystone -g keystone /etc/keystone/fernet-keys /etc/keystone/credential-keys
# Named volumes often mount as root:root; keystone-manage and Apache must read keys.
chown -R keystone:keystone /etc/keystone/fernet-keys /etc/keystone/credential-keys 2>/dev/null || true
mkdir -p /var/log/keystone
chown -R keystone:keystone /var/log/keystone
chmod 0775 /var/log/keystone
# Package default can leave a root-owned log path; keystone user must be able to write during db_sync.
rm -f /var/log/keystone/keystone-manage.log

if [ ! -f /etc/keystone/fernet-keys/0 ]; then
	keystone-manage fernet_setup --keystone-user keystone --keystone-group keystone
fi
if [ ! -f /etc/keystone/credential-keys/0 ]; then
	keystone-manage credential_setup --keystone-user keystone --keystone-group keystone
fi

echo "Running keystone-manage db_sync..."
# Run as root: packaged defaults still open /var/log/keystone/keystone-manage.log in some paths, and
# su keystone fails on a fresh image. Apache WSGI continues to run as the package user.
keystone-manage db_sync

USER_COUNT="0"
# Table name `user` must be quoted (reserved word). DB name is the last arg before -e.
USER_COUNT=$(mysql -h "$DB_HOST" -N -s -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" -e 'SELECT COUNT(*) FROM `user`' 2>/dev/null || echo 0)
if [ "${USER_COUNT:-0}" = "0" ]; then
	echo "Bootstrapping admin (first run)..."
	keystone-manage bootstrap \
		--bootstrap-password "$ADMIN_PASS" \
		--bootstrap-admin-url "$PUBLIC_URL" \
		--bootstrap-internal-url "$PUBLIC_URL" \
		--bootstrap-public-url "$PUBLIC_URL" \
		--bootstrap-region-id RegionOne
else
	echo "Keystone DB already has users; skipping bootstrap."
fi

exec apache2ctl -D FOREGROUND
