#!/bin/sh

ENV_FILE="$1"
CMD=${@:2}

source $ENV_FILE
export TOKEN=$TOKEN

$CMD