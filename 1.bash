#!/bin/bash
for i in {1..10}
do
   go test -v -test.run="TestLinuxMqSendIntSameProcess|XXX" || { echo "command failed"; exit 1; }
done 
