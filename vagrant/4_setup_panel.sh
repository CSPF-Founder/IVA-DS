#!/bin/bash
#mkdir -p /app/iva/infra/

# cd /app/build/panel/panelfiles
# make build_frontend
# make build

cd /app/build/panel/
sudo make up

mkdir -p /app/iva/panel/frontend/external


sudo cp -r /app/build/panel/panelfiles/frontend/external/* /app/iva/panel/frontend/external


sudo chown -R vagrant:vagrant /app/iva/