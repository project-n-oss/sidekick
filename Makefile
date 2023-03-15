all: sidekick

bin/staticcheck: go.mod go.sum
	GOBIN=`pwd`/bin go install honnef.co/go/tools/cmd/staticcheck

bin: bin/staticcheck 

sidekick:
	go build .

clean-bin:
	rm -rf bin
