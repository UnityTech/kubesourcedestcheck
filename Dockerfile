FROM alpine

ADD kubesourcedestcheck /
CMD /kubesourcedestcheck

EXPOSE 8061
# BUILD: # docker run --rm -v $PWD:/usr/src/$(basename $PWD) -w /usr/src/$(basename $PWD) golang:latest /bin/bash -c "go get -v ./...; CGO_ENABLED=0 go build -a;! ldd $(basename $PWD) && echo 'Build successful'"
