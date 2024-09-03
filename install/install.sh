#!/bin/sh

# Usage:
#   curl -sfL https://www.github.com/nhc-hvac/install.sh | sh -

# --- use sudo if we are not already root ---
SUDO=sudo
if [ "$(id -u)" -eq 0 ]; then
    SUDO=
fi

if [ -s /lib/systemd/system/nhcservice.service ]; then
  echo 'NHC service already installed -> stop and update'
  $SUDO systemctl stop nhcservice
fi

echo 'Downloading nhc from server'
wget https://www.github.com/nhc-hvac/nhcService
$SUDO chmod +x nhcService
$SUDO mkdir -p /usr/local/nhc
$SUDO mv nhcService /usr/local/nhc

echo 'Adding service user'
id -u nhcservice >/dev/null 2>&1 || $SUDO useradd nhcservice -s /sbin/nologin -M

echo 'Creating systemd service'
cat <<EOF | $SUDO tee /lib/systemd/system/nhcservice.service > /dev/null
[Unit]
Description=NHC Service
ConditionPathExists=/usr/local/nhc/nhcService
After=network.target
StartLimitIntervalSec=60

[Service]
Type=simple
User=nhcservice
Group=nhcservice
LimitNOFILE=1024

Restart=on-failure
RestartSec=10

WorkingDirectory=/usr/local/nhc
ExecStart=/usr/local/nhc/nhcService

# Output to journal to limit memory wear on RevPi
# journald is configured to store to tmpfs (RAM-based)
# PermissionsStartOnly=true #Deprecated, use +, ! and !!
# ExecStartPre=/bin/mkdir -p /var/log/nhcservice
# ExecStartPre=/bin/chown syslog:adm /var/log/nhcservice
# ExecStartPre=/bin/chmod 755 /var/log/nhcservice
# StandardOutput=syslog
# StandardError=syslog

# Also used by journal
SyslogIdentifier=nhcservice

[Install]
WantedBy=multi-user.target
EOF

echo 'Enabling NHC service'
$SUDO systemctl enable nhcservice.service

echo 'Starting NHC service'
$SUDO systemctl start nhcservice

echo ' '
echo 'Installation done'
echo 'See "sudo systemctl status nhc.service" and "journalctl -xe" for details'
echo 'To limit memory wear on the RevPi Core module, logs are volatile (RAM-based)'
echo 'See "sudo journalctl -f -u nhc" or "journalctl -u nhc" for all the journal'