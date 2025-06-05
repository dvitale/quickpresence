#!/bin/bash
for filename in $( ls -1 *html)
do
	awk 'BEGIN {print "<script> var filename=\"'$filename'\";</script>\n"}1' $filename  > $filename.new
	rm -f $filename
	mv $filename.new $filename


#	echo  "<!-- $filename -->" >> $filename  

done
