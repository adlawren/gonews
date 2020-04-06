FROM golang:1.14.1

# Disable terminal bell
RUN sed -i 's/# set bell-style none/set bell-style none/' /etc/inputrc

WORKDIR /go/src/gonews

COPY . .

RUN bash ./script/bootstrap
RUN bash ./script/setup

CMD ["bash", "./script/server"]
