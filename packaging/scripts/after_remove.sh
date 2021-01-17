#!/bin/sh

# parts of this file were taken from
# the do-agent packaging scripts located
# at https://github.com/digitalocean/do-agent/blob/master/packaging/scripts/after_remove.sh.

set -ue

SERVICE="fnd"

# fix an issue where this script runs on upgrades for rpm
# see https://github.com/jordansissel/fpm/issues/1175#issuecomment-240086016
arg="${1:-0}"

main() {
	if echo "${arg}" | grep -qP '^upgrade$'; then
		# deb upgrade
		exit 0
	fi

  if command -v systemctl >/dev/null 2>&1; then
    clean_systemd
  else
    echo "Unknown init system. Exiting." > /dev/stderr
    exit 1
  fi
}

clean_systemd() {
  echo "Cleaning up systemd scripts"
  systemctl stop ${SERVICE} || true
	systemctl disable ${SERVICE}.service || true
	systemctl daemon-reload || true
}

main
