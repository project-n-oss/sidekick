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

cat > /opt/spark/conf/style-path-spark-conf.conf <<EOL
[driver] {
  "spark.hadoop.fs.s3a.path.style.access" = "true"
  "spark.hadoop.fs.s3a.bucket.<MY_BUCKET1>.endpoint" = "http://localhost:7075"
  "spark.hadoop.fs.s3a.bucket.<MY_BUCKET1>.endpoint.region" = <AWS_REGION_OF_BUCKET1>
  "spark.hadoop.fs.s3a.bucket.<MY_BUCKET2>.endpoint" = "http://localhost:7075"
  "spark.hadoop.fs.s3a.bucket.<MY_BUCKET2>.endpoint.region" = <AWS_REGION_OF_BUCKET2>
}
EOL

# Add any spark or env config here:
# --------------------------------------------------

# --------------------------------------------------

export SIDEKICK_APP_CLOUDPLATFORM="<AWS|GCP>"

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
systemctl status sidekick
