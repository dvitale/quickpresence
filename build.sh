#!/bin/bash

FILE=$(basename $PWD)

if [ -f "./bin/$FILE" ]; then
   echo mv -f --backup=numbered "./bin/$FILE" "./old/$FILE"
   mv -f --backup=numbered "./bin/$FILE" "./old/$FILE"
fi

cd app
go build  -o "../bin/$FILE"
cd ..
##./stop.sh

