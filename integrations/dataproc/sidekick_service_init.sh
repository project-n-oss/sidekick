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

git clone -b feat/dataproc-integration-docs https://github.com/project-n-oss/sidekick.git /home/dataproc/sidekick
openssl x509 -outform der -in /home/dataproc/sidekick/cmd/sidekick-local.granica.ai.pem -out /home/dataproc/sidekick-local.granica.ai.der
keytool -noprompt -import -alias sidekickCert -keystore $JAVA_HOME/lib/security/cacerts -file /home/dataproc/sidekick-local.granica.ai.der -storepass changeit

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

export GRANICA_CUSTOM_DOMAIN="kmnov8.bolt.projectn.co"
export GRANICA_CLOUD_PLATFORM="gcp"

# Create service file for the sidekick process
SERVICE_FILE="/etc/systemd/system/sidekick.service"
touch $SERVICE_FILE

cat > $SERVICE_FILE << EOF
[Unit]
Description=Sidekick service file

[Service]
Environment=GRANICA_CUSTOM_DOMAIN=$GRANICA_CUSTOM_DOMAIN
Environment=GRANICA_CLOUD_PLATFORM=$GRANICA_CLOUD_PLATFORM
ExecStart=$SIDEKICK_BIN serve --cloud-platform gcp --log-level debug
Restart=always

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable sidekick
systemctl start sidekick
