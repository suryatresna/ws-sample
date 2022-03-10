FROM golang:1.16-alpine

# set working directory
WORKDIR /src
# copy all directory
ADD . .
# tun applicaiton
RUN go install ./main.go
# build the binary
ENTRYPOINT [ "main" ]