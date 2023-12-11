FROM golang:1.20.8-bullseye as builder

ENV GOPROXY=https://goproxy.io,direct

WORKDIR /
COPY . .
# COPY plugins/taos /usr/local/taos
EXPOSE  80

SHELL ["/bin/bash","-c"]
RUN bash build.sh common-dev /common-dev

ENTRYPOINT [ "/common-dev" ]
CMD ["common-dev"]
