#!/usr/bin/fish
make all
gh release create v0.0.9 ./dist/moralink-gost-linux-amd64 ./dist/moralink-gost-windows-amd64.exe ./dist/moralink-setup.exe ./dist/moralink-setup-x86.exe