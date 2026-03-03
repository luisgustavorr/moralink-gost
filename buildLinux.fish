#!/usr/bin/fish
make linux
sudo ./dist/moralink-gost-linux-amd64 uninstall
sudo ./dist/moralink-gost-linux-amd64 install
 systemctl restart moralink-gost.service