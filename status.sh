#!/bin/bash

FILE=$(basename $PWD)

echo -e "\nps:"
ps aux|grep $FILE |grep -v grep 
\
echo -e  "\nnetstat:"
netstat -natp |grep $(pidof $FILE)

echo ""
