#/bin/bash

for i in `seq 1 30`;
do
  nc -z $1 $2 && echo Success && exit 0
  echo -n .
  sleep 1
done
echo Failed waiting for $1 $2 && exit 1
