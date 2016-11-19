
FROM alpine
COPY loginsrv /
ENTRYPOINT ["/loginsrv"]
EXPOSE 6789
