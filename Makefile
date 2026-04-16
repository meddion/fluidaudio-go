.PHONY: all swift build test test-integration examples clean

all: swift build

swift: lib/release/libFluidAudioBridge.a

lib/release/libFluidAudioBridge.a: swift/FluidAudioBridge.swift Package.swift
	swift build -c release --build-path lib

build: swift
	go build ./...

test: swift
	go test -v -count=1 .

test-integration: swift
	go test -v -count=1 -tags integration .

examples: swift
	mkdir -p ./examples/bin && go build -o ./examples/bin ./examples/...

clean:
	rm -rf lib
