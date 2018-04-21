FROM alpine

RUN set -x \
  && apk add --update --no-cache ca-certificates

COPY webhook /usr/bin/webhook

ENTRYPOINT ["webhook"]
CMD ["run"]
EXPOSE 80
