#!/bin/bash
cd /app/build/manager/
make build
mkdir -p /app/iva/manager/
mkdir -p /app/iva/manager/local_temp/

cp /app/build/manager/bin/manager /app/iva/manager/manager
chmod +x /app/iva/manager/manager

cp /app/build/manager/.env /app/iva/manager/.env


sudo cp /app/build/manager/iva-manager.service /etc/systemd/system/iva-manager.service
sudo systemctl daemon-reload
sudo systemctl enable iva-manager.service
sudo systemctl start iva-manager.service

