#!/usr/bin/fish
make all SHARK_TOKEN=deploy
gh release create v0.1.0 ./clients_setups/deploy/moralink-gost-linux-amd64 ./clients_setups/deploy/moralink-gost-windows-amd64.exe ./clients_setups/deploy/moralink-setup.exe ./clients_setups/deploy/moralink-setup-x86.exe