#######
# Armillaria Docker 
#######

FROM golang:1.4.2

MAINTAINER Oslo Public Library "digitalutvikling@gmail.com"

ENV REFRESHED_AT 2015-05-25

RUN DEBIAN_FRONTEND=noninteractive apt-get update && \
    apt-get install -y wget less curl nmap netcat tree htop \ 
                       unzip sudo python-software-properties  && \
    apt-get clean

ENV SERVER_PORT         8080
ENV SPARUL_ENDPOINT     http://localhost:8890/sparql-auth
ENV SPARUL_USER         dba
ENV SPARUL_PASS         dba
ENV ELASTICSEARCH_URL   http://localhost:9200
ENV DEFAULT_GRAPH       http://example.com/test

ENV HOME /app
WORKDIR /app

ADD . /app

RUN cd /app && make deps && make build

EXPOSE 8080 9200 9300

CMD /app/armillaria