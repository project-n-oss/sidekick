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


# Setup spark hadoop s3 configuration 
# https://docs.databricks.com/storage/amazon-s3.html#global-configuration
# TODO: change <MY_BUCKET> to your crunched bucket
# If you want access to multiple buckets, just add a new config line for every crunched bucket you want access to.
cat >/databricks/driver/conf/sidekick-spark-conf.conf <<EOL
[driver] {
  "spark.hadoop.fs.s3a.path.style.access" = "true"
  "spark.hadoop.fs.s3a.bucket.<MY_BUCKET>.endpoint" = "http://localhost:8081"
}
EOL


# Create service file for the sidekick process
# TODO: Change <YOUR_CUSTOM_BOLT_DOMAIN> to your custom bolt domain
export BOLT_CUSTOM_DOMAIN=<YOUR_CUSTOM_BOLT_DOMAIN>
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
