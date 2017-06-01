docker run --rm -v ${PWD}:/go/src/github.com/UnityTech/kubesourcedestcheck -w /go/src/github.com/UnityTech/kubesourcedestcheck golang /bin/bash -c "curl https://glide.sh/get | sh && glide install && CGO_ENABLED=0 go build -a --installsuffix cgo -ldflags='-s -w -X main.builddate=`date -u +\"%Y-%m-%dT%H:%M:%SZ\"`' -v && ! ldd kubesourcedestcheck"

docker build -t $image .
