#!/bin/bash

logRoot="./log"
configFile="../../bin/config/config.json"

go build -o logicserver
if [ $? -eq 0 ]; then
    echo "logicserver compile success"
    ps aux|grep "logicserver"|sed -e "/grep/d"|awk '{print $2}'|xargs kill -9 2&>/dev/null
    if [ ! -d "$logRoot" ]; then
        mkdir "$logRoot"
    else
        echo "$logRoot"" already existed"
    fi
    for port in $*
    do
        logPath=$logRoot"/logfile_"$port"/"
        if [ ! -d "$logPath" ]; then
            mkdir "$logPath"
        fi
        ./logicserver -config "$configFile" -port "$port" -log_dir="$logPath" -alsologtostderr &
    done
else
    echo "logicserver compile failed"
fi