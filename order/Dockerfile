FROM alpine:3.12.0
RUN apk add --update ca-certificates && \
    rm -rf /var/cache/apk/* /tmp/*
ENV PORT 2048
ENV SERVER_NAME order.Order
ENV SERVER_ADDRESS 0.0.0.0:2048
WORKDIR /app
ADD _bin/${SERVER_NAME} /app
EXPOSE ${PORT}
CMD [ "./order" ]
