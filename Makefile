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