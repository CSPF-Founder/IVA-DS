#!/bin/bash
export DEBIAN_FRONTEND=noninteractive

# Permission set
mkdir -p /app/
sudo chown -R vagrant:vagrant /app
#Change DNS
sudo apt-get install resolvconf -y
echo "nameserver 8.8.8.8" >> sudo tee -a /etc/resolvconf/resolv.conf.d/base
echo "nameserver 8.8.4.4" >> sudo tee -a /etc/resolvconf/resolv.conf.d/base
sudo resolvconf -u
# Add Docker's official GPG key:
sudo apt-get update -y 
sudo apt-get install ca-certificates curl -y
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

# Add the repository to Apt sources:
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update

sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin -y

#Setup Docker Environment
sudo usermod -aG docker $USER
sudo usermod -aG docker vagrant
sudo gpasswd -a $USER docker
sudo gpasswd -a vagrant docker
