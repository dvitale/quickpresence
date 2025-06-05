#!/bin/bash

clear

#set -e

set -x

./stop.sh 
 ./build.sh 
 ./run.sh 
#clear
>nohup.out
 tail -f nohup.out
