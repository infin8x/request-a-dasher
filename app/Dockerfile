FROM golang:alpine as builder

RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go build -o main .

FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /build/main /app/
COPY /templates /app/templates
COPY /static /app/static
WORKDIR /app
EXPOSE 8080
CMD ["./main"]