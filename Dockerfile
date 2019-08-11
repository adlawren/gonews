FROM golang:1.12

WORKDIR /go/src/gonews

COPY . .

EXPOSE 8080

# Disable terminal bell # TODO: test? rm?
#RUN rmmod pcspkr

RUN go clean
RUN go get -d -v ./...
RUN go install -v ./...

ENTRYPOINT ["./docker/run.sh"]
