#!/bin/bash

FILE=$(basename $PWD)

echo " kill -9 $(pidof $FILE) "
 kill -9 $(pidof $FILE)

