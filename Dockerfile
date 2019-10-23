FROM golang:1.12

WORKDIR /go/src/gonews

COPY . .

EXPOSE 8080

# Disable terminal bell
RUN sed -i 's/# set bell-style none/set bell-style none/' /etc/inputrc

RUN go clean
RUN go get -d -v ./...
RUN go install -v ./...

ENTRYPOINT ["./docker/run.sh"]
