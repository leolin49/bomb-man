#!/bin/bash

logPath="./log/"
configFile="../../bin/config/config.json"
serverName="roomserver"

go build -o "$serverName"
if [ $? -eq 0 ];then
    echo "$serverName""compile success"
    if [ ! -d "$logPath" ];then
        mkdir "$logPath"
    else
        echo "$logPath""already exsit"
    fi
    ./"$serverName" -config "$configFile" -log_dir="$logPath" -alsologtostderr
else
    echo "$serverName""compile failed"
fi