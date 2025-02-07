#!/bin/bash

cd /app/build/scanner/

make build

cp /app/build/scanner/bin/scanner /app/iva/scanner/scanner

chmod +x /app/iva/scanner/scanner

cp /app/build/scanner/.env /app/iva/scanner/.env


echo '#!/bin/bash' > /app/iva/bin/scanner && echo 'cd /app/iva/scanner/ && ./scanner $@' >> /app/iva/bin/scanner

chmod +x /app/iva/bin/scanner