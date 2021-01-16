#!/bin/sh

# parts of this file were taken from
# the do-agent packaging scripts located
# at https://github.com/digitalocean/do-agent/blob/master/packaging/scripts/after_install.sh.

set -ue

USERNAME="ddrp"
SERVICE="ddrpd"
UNIT_FILE="/etc/systemd/system/${SERVICE}.service"

main() {
  echo "Adding ${USERNAME} user..."
  useradd -s /bin/false -M --system $USERNAME
  echo "Initializing ddrpd config in /etc/ddrpd.."
  /usr/bin/ddrpd init --home /etc/ddrpd || true
  chown -R $USERNAME:$USERNAME /etc/ddrpd

  if command -v systemctl >/dev/null 2>&1; then
    rm -f "${UNIT_FILE}"
    echo "Enabling systemd service..."
    systemctl daemon-reload
    systemctl enable -f "${SERVICE}"
    systemctl restart "${SERVICE}"
  else
    echo "Unknown init system. Exiting." > /dev/stderr
    exit 1
  fi

  echo "Done."
}

main