#!/bin/bash

### DV 2021

DOWNNODES=$(grep -iPo '"status":.*?[^\\]",' /opt/go_projects/gomo/bin/nodes.json |grep -i DOWN |wc -l)

PIDOFGOMO=$(pidof gomo)

MSGNODES="Attenzione alcuni nodi potrebbero essere DOWN"
MSGAPP="Attenzione gomo non risulta essere in esecuzione" 

DEST=gisahelpdesk@u-s.it

if [[ -z "$PIDOFGOMO" ]] ; then
	echo $MSGAPP #| mailx -s "gomo WARNING"  -r gomo@u-s.it $DEST
elif [[ "$DOWNNODES" -gt 0 ]] ; then
	echo $MSGNODES #| mailx -s "gomo WARNING"  -r gomo@u-s.it $DEST
fi




