#!/bin/sh

# Usage:
#   curl -sfL https://www.couillonnade.com/revpi/install.sh | sh -

# --- use sudo if we are not already root ---
SUDO=sudo
if [ "$(id -u)" -eq 0 ]; then
    SUDO=
fi

if [ -s /lib/systemd/system/nbaservice.service ]; then
  echo 'nba service already installed -> stop and update'
  $SUDO systemctl stop nbaservice
fi

echo 'Downloading nba from server'
wget https://www.couillonnade.com/revpi/nbaService
$SUDO chmod +x nbaService
$SUDO mkdir -p /usr/local/nba
$SUDO mv nbaService /usr/local/nba

echo 'Adding service user'
id -u nbaservice >/dev/null 2>&1 || $SUDO useradd nbaservice -s /sbin/nologin -M

echo 'Creating systemd service'
cat <<EOF | $SUDO tee /lib/systemd/system/nbaservice.service > /dev/null
[Unit]
Description=NBA Service
ConditionPathExists=/usr/local/nba/nbaService
After=network.target
StartLimitIntervalSec=60

[Service]
Type=simple
User=nbaservice
Group=nbaservice
LimitNOFILE=1024

Restart=on-failure
RestartSec=10

WorkingDirectory=/usr/local/nba
ExecStart=/usr/local/nba/nbaService

# Output to journal to limit memory wear on RevPi
# journald is configured to store to tmpfs (RAM-based)
# PermissionsStartOnly=true #Deprecated, use +, ! and !!
# ExecStartPre=/bin/mkdir -p /var/log/nbaservice
# ExecStartPre=/bin/chown syslog:adm /var/log/nbaservice
# ExecStartPre=/bin/chmod 755 /var/log/nbaservice
# StandardOutput=syslog
# StandardError=syslog

# Also used by journal
SyslogIdentifier=nbaservice

[Install]
WantedBy=multi-user.target
EOF

echo 'Enabling NBA service'
$SUDO systemctl enable nbaservice.service

echo 'Starting NBA service'
$SUDO systemctl start nbaservice

echo ' '
echo 'Installation done'
echo 'See "sudo systemctl status nbaservice.service" and "journalctl -xe" for details'
echo 'To limit memory wear on the RevPi Core module, logs are volatile (RAM-based)'
echo 'See "sudo journalctl -f -u nbaservice" or "journalctl -u nbaservice" for all the journal'