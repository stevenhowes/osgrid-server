#!/bin/bash
TOKEN=`uuidgen`
echo $TOKEN
MD5=`echo -n $TOKEN  | md5sum | awk '{ print $1 }'`
echo $MD5
echo -n 0 > keys/$MD5
