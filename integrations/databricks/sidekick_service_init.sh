#!/bin/bash
set -ex

# Check if sidekick bin is present, if not download it
SIDEKICK_BIN=/usr/bin/sidekick
if [ -f "$SIDEKICK_BIN" ]; then
    echo "$SIDEKICK_BIN already installed."
else 
    curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
    unzip awscliv2.zip
    sudo ./aws/install
    aws s3 cp s3://sidekick-test-rvh2/bin/sidekick /usr/bin/sidekick
fi
chmod +x /usr/bin/sidekick

cat >/databricks/driver/conf/style-path-spark-conf.conf <<EOL
[driver] {
  "spark.hadoop.fs.s3a.path.style.access" = "true"
}
EOL

# Add any spark or env config here:
# --------------------------------------------------

# cat >/databricks/driver/conf/sidekick-spark-conf.conf <<EOL
# [driver] {
#   "spark.hadoop.fs.s3a.bucket.<MY_BUCKET>.endpoint" = "http://localhost:7071"
# }
# EOL

# export BOLT_CUSTOM_DOMAIN=<YOUR_CUSTOM_BOLT_DOMAIN>

# --------------------------------------------------

# Create service file for the sidekick process
SERVICE_FILE="/etc/systemd/system/sidekick.service"
touch $SERVICE_FILE

cat > $SERVICE_FILE << EOF
[Unit]
Description=Sidekick service file

[Service]
Environment=BOLT_CUSTOM_DOMAIN=$BOLT_CUSTOM_DOMAIN
ExecStart=/usr/bin/sidekick serve
Restart=always

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable sidekick
systemctl start sidekick


