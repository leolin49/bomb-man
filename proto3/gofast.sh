#!/bin/bash

protoc --gofast_out=plugins=grpc:. $1