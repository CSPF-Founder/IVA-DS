#!/bin/bash

mkdir -p /app/iva/reporter/src

cp -r /vagrant/code/reporter/src/. /app/iva/reporter/src


cd /app/iva/reporter/src
python3 -m venv venv
source venv/bin/activate
pip install poetry
poetry install --no-interaction --no-root

# GVM cli for openvas connection
pip install gvm-tools==24.3.0
pip install python-gvm==24.3.0
deactivate

# Create Reporter bin shell

echo '#!/bin/bash' > /app/iva/bin/reporter && echo 'cd /app/iva/reporter/src && source venv/bin/activate && python3 cli.py $@' >> /app/iva/bin/reporter

chmod +x /app/iva/bin/reporter

ln -s /app/iva/reporter/src/venv/bin/gvm-cli /app/iva/bin/gvm-cli