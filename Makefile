# go get github.com/akavel/rsrc

build:
	rsrc -manifest manifest -ico icon.ico -o rsrc.syso
	go build -ldflags="-H windowsgui"

dist: build
	"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" setup.iss

download:
	curl -LO# https://github.com/Genymobile/scrcpy/releases/download/v1.17/scrcpy-win64-v1.17.zip
	unzip -d scrcpy scrcpy-win64-v1.17.zip

.PHONY: build dist
