#!/usr/bin/env fish

mkdir -p assets

pushd assets

curl -s https://github.com/icub3d/mdlp | \
	rg '<link' | rg 'rel="stylesheet"' | \
	rg -o 'href="[^"]*"' | \
	sed -e 's/"//g' -e 's/href=//g' | \
	xargs -n1 wget

popds
