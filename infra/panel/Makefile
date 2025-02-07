DOCKER_STACK = iva

setup: create_dirs cert
	@echo "First time setup done(dont call this agin)!"

create_dirs:
	@echo "First time setup..."
	mkdir -p /app/iva/panel/certs/
	mkdir -p /app/iva/panel/frontend/external/
	mkdir -p /app/iva/data/temp_uploads/
	mkdir -p /app/iva/data/work_dir/
	mkdir -p /app/iva/bin/
	mkdir -p /app/iva/manager/local_temp/
	mkdir -p /app/iva/scanner/local_temp/
	mkdir -p /app/iva/reporter/src
	mkdir -p /app/iva/reporter/config
	
	mkdir -p /app/iva/data/reports/
	mkdir -p /app/iva/logs/scans/


cert:
	@echo "Generating SSL Certificate..."
	openssl req -x509 -sha256 -nodes -days 4096 -newkey rsa:2048 -keyout /app/iva/panel/certs/panel.key -out /app/iva/panel/certs/panel.crt -subj "/C=US/ST=CSPF/L=CSPF/O=Dis/CN=iva"
	@echo "SSL Generated"

up:
	@echo "Starting containers..."
	docker compose -p ${DOCKER_STACK} up -d
	@echo "Containers started!"

#up_build: starts all containers in the background and recreates them
up_build:
	@echo "Stopping containers..."
	docker compose -p ${DOCKER_STACK} down
	@echo "Building (when required) and Starting containers..."
	docker compose -p ${DOCKER_STACK} up -d --build
	@echo "Docker images build and containers started!"

# down: stops all containers
down:
	@echo "Stopping containers..."
	docker compose -p ${DOCKER_STACK} down
	@echo "Containers stopped!"
