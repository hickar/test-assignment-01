FROM golang:1.22-alpine as build

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o ./order-service ./order/cmd/


FROM alpine 

WORKDIR /app

COPY --from=build /app/order-service .

ENTRYPOINT ["/app/order-service"]
