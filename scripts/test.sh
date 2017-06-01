docker run --rm -v ${PWD}:/go/src/github.com/UnityTech/kubesourcedestcheck -w /go/src/github.com/UnityTech/kubesourcedestcheck golang /bin/bash -c "go get -v github.com/kardianos/govendor && govendor sync && go test -v "

