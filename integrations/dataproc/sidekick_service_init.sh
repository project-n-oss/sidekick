#!/bin/bash
set -ex

# Check if sidekick bin is present, if not download it
SIDEKICK_BIN=/usr/bin/sidekick
if [ -f "$SIDEKICK_BIN" ]; then
    echo "$SIDEKICK_BIN already installed."
else 
    wget https://github.com/project-n-oss/sidekick/releases/latest/download/sidekick-linux-amd64.tar.gz
    tar -xzvf sidekick-linux-amd64.tar.gz -C /usr/bin
fi
chmod +x $SIDEKICK_BIN
$SIDEKICK_BIN --help > /dev/null

sed -i '/<\/configuration>/i \
  <property>\
    <name>fs.gs.storage.root.url</name>\
    <value>http://localhost:7075</value>\
    <description>\
      Google Cloud Storage root URL.\
    </description>\
  </property>' /etc/hadoop/conf/core-site.xml

# Add any spark or env config here:
# --------------------------------------------------

# --------------------------------------------------

export GRANICA_CUSTOM_DOMAIN=<YOUR_CUSTOM_DOMAIN>
export GRANICA_CLOUD_PLATFORM="<aws|gcp>"
export SIDEKICK_EXTRA_ARGS="<CHECK MAIN README FOR OTHER ARGS"

# Create service file for the sidekick process
SERVICE_FILE="/etc/systemd/system/sidekick.service"
touch $SERVICE_FILE

cat > $SERVICE_FILE << EOF
[Unit]
Description=Sidekick service file

[Service]
Environment=GRANICA_CUSTOM_DOMAIN=$GRANICA_CUSTOM_DOMAIN
Environment=GRANICA_CLOUD_PLATFORM=$GRANICA_CLOUD_PLATFORM
Environment=SIDEKICK_EXTRA_ARGS=$SIDEKICK_EXTRA_ARGS
ExecStart=$SIDEKICK_BIN serve --cloud-platform $GRANICA_CLOUD_PLATFORM $SIDEKICK_EXTRA_ARGS
Restart=always

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable sidekick
systemctl start sidekick
