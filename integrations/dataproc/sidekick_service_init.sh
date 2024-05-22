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

export SIDEKICK_APP_CLOUDPLATFORM="<aws|gcp>"

# Create service file for the sidekick process
SERVICE_FILE="/etc/systemd/system/sidekick.service"
touch $SERVICE_FILE

cat > $SERVICE_FILE << EOF
[Unit]
Description=Sidekick service file

[Service]
Environment=SIDEKICK_APP_CLOUDPLATFORM=$SIDEKICK_APP_CLOUDPLATFORM
ExecStart=$SIDEKICK_BIN serve -p 7075
Restart=always

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable sidekick
systemctl start sidekick
