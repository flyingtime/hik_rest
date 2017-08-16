#!/bin/bash

echo "/hkrest" >> /etc/ld.so.conf
echo "/hkrest/HCNetSDKCom/" >> /etc/ld.so.conf
ldconfig
cd hkrest
./main
