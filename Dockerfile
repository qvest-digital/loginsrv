FROM alpine
RUN apk --update add ca-certificates \
    && addgroup -S loginsrv && adduser -S -g loginsrv loginsrv
USER loginsrv
ENV LOGINSRV_HOST=0.0.0.0 LOGINSRV_PORT=8080
COPY loginsrv /
ENTRYPOINT ["/loginsrv"]
EXPOSE 8080
