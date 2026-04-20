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

example-diarize: swift
	go run ./examples/diarize/main.go testdata/test.wav

example-diarize-stream: swift
	go run ./examples/streaming-diarize/main.go testdata/test.wav

clean:
	rm -rf lib
