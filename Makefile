help:
	@echo "help"
	@echo "build"
	@echo "install"
	@echo "release"

.PHONY: build
build:
	@go build -o qn_cli main.go

.PHONY: install
install:
	@go install github.com/mozillazg/qn_cli

.PHONY: release
release:
	@goxc -d=build -pv=`cat version.txt` -bc='linux,windows,darwin'
