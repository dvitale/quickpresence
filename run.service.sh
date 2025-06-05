#!/bin/bash


cd $(/usr/bin/dirname $0)
 

FILE=$(/usr/bin/basename $PWD)

#echo "$FILE start: $(date)">>exec.log 
bin/${FILE} >>./log/out.txt 2>>./log/err.txt  
#echo "$FILE end  : $(date)" >>exec.log
