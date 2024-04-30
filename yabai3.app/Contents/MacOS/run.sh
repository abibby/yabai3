#!/usr/bin/env bash

cd $(dirname $0)

./yabai3 >> "$HOME/Library/Logs/yabai3.log" 2>&1
