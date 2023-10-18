.DEFAULT_GOAL := build

clean:
	rm -f osgrid-server

fmt:
	go fmt ./...
.PHONY:fmt

lint: fmt
	golint ./...
.PHONY:lint

vet: fmt
	go vet ./...
.PHONY:vet

build: vet
	go build
.PHONY:build

run: build
	./osgrid-server
.PHONY:run

install: build
	mkdir -p /opt/sartools/
	useradd sartools || true
	chown sartools:sartools /opt/sartools/
	cp osgrid-server /opt/sartools/
	cp osgrid-server.service /etc/systemd/system/
