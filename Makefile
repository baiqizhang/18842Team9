default:
	go build -o bin/carnode dsproject/carnode
	go build -o bin/server dsproject/server
	go build -o bin/supernode dsproject/supernode
clean:
	rm -rf bin/*
