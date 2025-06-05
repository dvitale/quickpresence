#!/bin/bash


cd $(dirname $0)
 

FILE=$(basename $PWD)

#echo "$FILE start: $(date)">>exec.log 
nohup bin/${FILE} >>./log/out.txt 2>>./log/err.txt  &
#echo "$FILE end  : $(date)" >>exec.log
