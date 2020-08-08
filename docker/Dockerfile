FROM golang:1.14.1

WORKDIR /go/src/gonews

COPY . .

RUN bash ./script/bootstrap

CMD ["bash", "./script/server"]