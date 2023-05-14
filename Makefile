version=$(shell git describe --always --long --dirty)
date=$(shell TZ=UTC date)
commit=$(shell git log -1 --pretty=format:"%H")

genCert:
	cd cert && ./gen.sh

client:
	go build -o keeperClient -ldflags '\
		-X "main.buildVersion=${version}"\
		-X "main.buildDate=${date}"\
		-X "main.buildCommit=${commit}"' \
	  cmd/client/main.go
