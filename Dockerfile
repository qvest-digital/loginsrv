FROM alpine
RUN addgroup -S loginsrv && adduser -S -g loginsrv loginsrv
USER loginsrv
ENV LOGINSRV_HOST=0.0.0.0 LOGINSRV_PORT=80
COPY loginsrv /
ENTRYPOINT ["/loginsrv"]
EXPOSE 80
