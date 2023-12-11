.PHONY: generate proto file

gen:
	protoc -I ./proto ./proto/frame.proto --go_out=plugins=grpc:./pb

clean:
	rm pb/*.go
	