#!/bin/bash
mkdir -p /app/build/panel/
mkdir -p /app/build/scanner/
mkdir -p /app/iva/reporter/config
mkdir -p /app/build/libs/
#mkdir -p /app/build/scanner-api/
mkdir -p /app/build/manager/



cp -a /vagrant/infra/panel/. /app/build/panel/
cp -a /vagrant/code/scanner/. /app/build/scanner/
cp -a /vagrant/code/libs/. /app/build/libs/
cp -r /vagrant/code/reporter/config/. /app/iva/reporter/config
#cp -a /vagrant/code/scanner-api/. /app/build/scanner-api/
cp -a /vagrant/code/manager/. /app/build/manager/



#Edit and make SSL panel
cd /app/build/panel/

rootpass=`(sudo head /dev/urandom | tr -dc 'A-Za-z0-9' | head -c 20)`
normalpass=`(sudo head /dev/urandom | tr -dc 'A-Za-z0-9' | head -c 20)`
gmppass=`(sudo head /dev/urandom | tr -dc 'A-Za-z0-9' | head -c 20)`

sed -i "s/\[ROOT_PASS_TO_REPLACE\]/$rootpass/g" /app/build/panel/docker-compose.yaml
sed -i "s/\[PASSWORD_TO_REPLACE\]/$normalpass/g" /app/build/panel/docker-compose.yaml
sed -i "s/\[GMP_PASSWORD_TO_REPLACE\]/$gmppass/g" /app/build/panel/docker-compose.yaml
sed -i "s/\[GMP_PASSWORD_TO_REPLACE\]/$gmppass/g" /app/build/scanner/env.example


sudo make


randomapikey=`(sudo head /dev/urandom | tr -dc 'A-Za-z0-9' | head -c 40)`
cp /app/build/scanner/env.example /app/build/scanner/.env
sed -i "s/\[ROOT_PASS_TO_REPLACE\]/$rootpass/g" /app/build/scanner/.env
sed -i "s/\[API_KEY\]/$randomapikey/g" /app/build/scanner/.env
sed -i "s/\[API_KEY\]/$randomapikey/g" /app/build/panel/docker-compose.yaml



#Edit reporter config
mkdir -p /app/iva/reporter/config
sed -i "s/\[ROOT_PASS_TO_REPLACE\]/$rootpass/g" /app/iva/reporter/config/app.conf


#Edit manager config
cp /app/build/manager/env.example /app/build/manager/.env
sed -i "s/\[ROOT_PASS_TO_REPLACE\]/$rootpass/g" /app/build/manager/.env
#chown back

sudo chown -R vagrant:vagrant /app/iva/