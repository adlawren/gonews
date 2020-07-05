FROM golang:1.14.1

WORKDIR /go/src/gonews

COPY . .

RUN bash ./script/bootstrap
RUN bash ./script/setup

CMD ["bash", "./script/server"]
