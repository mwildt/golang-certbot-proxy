FROM certbot/certbot

MAINTAINER mail@maltewildt.de

RUN mkdir -p /opt/mwcertbot
WORKDIR /opt/mwcertbot

COPY main /opt/mwcertbot/main

RUN mkdir -p /opt/mwcertbot/.well-known/
VOLUME /opt/mwcertbot/.well-known/

ENTRYPOINT ["/bin/sh", "-l", "-c"]
CMD ["/opt/mwcertbot/main"]