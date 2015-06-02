#######
# Armillaria Docker 
#######

FROM golang:1.4.2

MAINTAINER Oslo Public Library "digitalutvikling@gmail.com"

ENV REFRESHED_AT 2015-05-25

RUN DEBIAN_FRONTEND=noninteractive apt-get update && \
    apt-get install -y wget less curl netcat tree htop patch \
                       unzip sudo python-software-properties && \
    apt-get clean

ENV SERVER_PORT         8080
ENV SPARUL_USER         dba
ENV SPARUL_PASS         secret
ENV DEFAULT_GRAPH       http://example.com/test

ENV HOME /app
WORKDIR /app

##############
# DEPENDENCIES
##############

# go libs
RUN go get github.com/digibib/armillaria
RUN go get github.com/julienschmidt/httprouter
RUN go get github.com/knakk/sparql
RUN go get github.com/knakk/digest
RUN go get github.com/knakk/rdf

# html5 libs
RUN mkdir -p /app/data/public/css && \
    mkdir -p /app/data/public/js
RUN wget -q --no-check-certificate http://necolas.github.com/normalize.css/3.0.1/normalize.css -O /app/data/public/css/normalize.css
RUN wget -q --no-check-certificate http://cdn.ractivejs.org/0.7.3/ractive.js -O /app/data/public/js/ractive.js
RUN wget -q --no-check-certificate https://github.com/ractivejs/cdn.ractivejs.org/blob/gh-pages/0.7.3/ractive.min.js.map -O /app/data/public/js/ractive.js.map
RUN wget -q --no-check-certificate https://raw.github.com/ractivejs/ractive-events-keys/master/dist/ractive-events-keys.js -O /app/data/public/js/ractive-events-keys.js
RUN wget -q --no-check-certificate http://underscorejs.org/underscore-min.js -O /app/data/public/js/underscore-min.js

ADD . /app

# patch ractive.js for event.index
RUN wget -q --no-check-certificate \
  https://gist.githubusercontent.com/bensinober/414f0babd78efc052321/raw/9c633f6465c6b1ac8d4fec8512e1c76540982ee8/ractive-patch-gh-1982 -O - | patch -d /app/data/public/js

RUN go build -o armillaria

EXPOSE 8080

CMD /app/armillaria