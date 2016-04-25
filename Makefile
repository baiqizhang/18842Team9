all: ./bin/carnode ./bin/server ./bin/supernode

./bin/carnode: ./carnode/*
	go build -o bin/carnode dsproject/carnode

./bin/server: ./server/*
	go build -o bin/server dsproject/server

./bin/supernode: ./supernode/*
	go build -o bin/supernode dsproject/supernode

clean:
	rm -rf bin/*
