# IVA Scanner

## About

IVA Scanner is a vulnerability scanner that detects security weaknesses in networks and integrates with a bug tracker to streamline tracking and resolution. It performs differential scans on websites and IPs, prioritizing vulnerabilities and comparing results over time. Using ZAP for web scanning and OpenVAS for network scanning, IVAScanner provides comprehensive coverage and detects a wider range of threats.

## Key Features

- **Integrated Bug Tracker:** Easily track and manage vulnerabilities from discovery to resolution.
- **Powerful Scanning Engines:** Uses ZAP for web scans and OpenVAS for network scans to detect a wide range of threats.
- **Comprehensive Coverage:** Combines ZAP and OpenVAS to identify more vulnerabilities than standalone tools.
- **Efficient Workflow:** Enhances efficiency by simplifying vulnerability management, reducing risks, and saving time.

## Prerequisites

1. Install Vagrant from the official site, https://developer.hashicorp.com/vagrant/downloads. 

- Please refer to this Installation guide if you face any issues during installation. https://developer.hashicorp.com/vagrant/docs/installation
  

2. Install Virtualbox from the official site, https://www.virtualbox.org/wiki/Downloads

## Minimum Spec

- 8 GB RAM 
- 4 CPU cores
- 50 GB of free disk space.
  
You can adjust the disk space in the Vagrantfile (line 29). The minimum should be 25 GB.

## Installing VM

Download the repository via 

`git clone https://github.com/CSPF-Founder/iva.git`

Or you can download it as a zip file by clicking on `Code` in the top right and clicking `Download zip`.

`cd` into the folder that is created.

### In Linux:

In the project folder run the below command.

```
chmod +x setupvm.sh

./setupvm.sh
```

Once the vagrant installation is completed, it will automatically restart in Linux. 

### In Windows:

Go to the project folder on command prompt and then run the below commands.

```
vagrant up
```
After it has been completed, run the below command to reload the VM manually.

```
vagrant reload
```


## Accessing the Panel

The IVA Scanner Panel is available on this URL: https://localhost:8443. 

```
Note: If you want to change the port, you can change the forwardport in the vagrantfile (line 28).
```

For information on how to use the panel refer to [Manual.md](Manual.md)

## Further Reading:

- It is highly recommended to change the default password of the user `vagrant` and change the SSH keys. 

- If you want to start the VM after your computer restarts you can give `vargant up` on this folder or start from the virtualbox manager. 

- Once up you can access the VM by running the command `vagrant ssh` from the project folder.
