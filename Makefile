ROOT := ${CURDIR}

GCFLAGS='-N -l'
GCLDFLAGS=''
BUILD_CMD =go build \
                -gcflags=${GCFLAGS} \
                -ldflags=${GCLDFLAGS} \
                -o ${ROOT}/bin/$@ \
                ./cmd/$@

all: unit-test

unit-test:
	${BUILD_CMD}

clean:
	rm -f bin/*

realclean: clean
	go clean -modcache -cache
