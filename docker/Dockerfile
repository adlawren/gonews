FROM golang:1.12

# Disable terminal bell
RUN sed -i 's/# set bell-style none/set bell-style none/' /etc/inputrc

WORKDIR /go/src/gonews

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...