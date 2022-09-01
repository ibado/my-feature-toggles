#!/bin/bash

pid=-1
program_name=program

exitfn () {
    pkill -f $program_name
}

trap "exitfn" EXIT

run_server() {
    if [ $pid -ne -1 ]
    then
        kill $pid
    fi
    echo "executing go build & run..."
    echo ""
    go build -o /tmp/$program_name .
    /tmp/$program_name &
    pid=$!
}

run_server

inotifywait -q -r -e close_write -m . |
while read -r directory events filename; do
  if [[ $filename == *\.go ]]; then
    echo ""
    echo "Changes detected!"  
    run_server
  fi
done
