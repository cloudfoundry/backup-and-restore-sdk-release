FROM alpine:latest AS resource

RUN apk --no-cache add \
    bash \
    jq \
    wget \
    curl \
    tar \
    python3 \
    py3-pip

RUN pip3 install yq

WORKDIR /root

RUN wget -4 https://github.com/progrium/basht/releases/download/v0.1.0/basht_0.1.0_Linux_x86_64.tgz && \
    tar zxvf basht_0.1.0_Linux_x86_64.tgz && \
    install basht /usr/bin/basht && \
    rm basht

ADD . /opt/resource
RUN chmod +x /opt/resource/*

RUN cd /opt/resource && ./all_test.bash

FROM resource

WORKDIR /opt/resource