#!/bin/bash

# If not installed you have to isntall ImageMagick
# sudo apt install imagemagick

convert $1 $2
convert $2 -resize 640x $2
convert $2 -crop 640x480+0+0 $2