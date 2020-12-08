FROM alpine:3.12.0
RUN apk add --update ca-certificates && \
    rm -rf /var/cache/apk/* /tmp/*
WORKDIR /app
ADD _bin/health /app
CMD [ "./health" ]
