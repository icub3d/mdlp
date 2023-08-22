#!/usr/bin/env bash

if [ ! -d styles ]; then
	mkdir -p styles
	
	pushd styles

	curl -s https://github.com/icub3d/mdlp | \
		rg '<link' | rg 'rel="stylesheet"' | \
		rg -o 'href="[^"]*"' | \
		sed -e 's/"//g' -e 's/href=//g' | \
		xargs -n1 wget
	
	popd
fi
