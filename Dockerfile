FROM golang:1.16.13-alpine3.15 as build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -a -o kvs

FROM scratch
COPY --from=build /app/kvs .
COPY *.pem .
EXPOSE 8009
CMD ["/kvs"]