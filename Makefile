export CGO_ENABLED=1
export WDIR=${PWD}

all: linux

docker: linux
	sudo docker build -t xujieasd/hik-rest .

linux:
	GOOS=linux CGO_CFLAGS="-I${WDIR}/include" CGO_LDFLAGS="-L${WDIR}/lib -Wl,-rpath=${WDIR}/:${WDIR}/lib:${WDIR}/lib/HCNetSDKCom -lhcnetsdk" go build -ldflags "-s -w" -o build/main main.go
	cp lib/*.so build/
	cp -r lib/HCNetSDKCom/ build/

clean:
	rm -r build/