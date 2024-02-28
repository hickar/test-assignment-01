FROM golang:1.22-alpine as build

WORKDIR /app

ADD go.mod .
ADD go.sum .
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o ./orderservice ./cmd/main/


FROM alpine 

WORKDIR /app

COPY --from=build /app/orderservice /app/orderservice

CMD ["/app/orderservice"]
