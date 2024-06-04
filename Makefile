all: sidekick

bin/staticcheck: go.mod go.sum
	GOBIN=`pwd`/bin go install honnef.co/go/tools/cmd/staticcheck

bin: bin/staticcheck 

.PHONY: sidekick
sidekick: clean-bin
	go build .

default-pgo: cpu.prof
	cp cpu.prof default.pgo

clean-bin:
	rm -rf bin
