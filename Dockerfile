FROM --platform=linux/amd64 golang:1.22-alpine as build
WORKDIR /app

COPY go.sum go.mod ./
RUN go mod download

COPY . .

RUN go build -o main .

FROM --platform=linux/amd64 alpine
WORKDIR /app

COPY --from=build /app/main /app/main

CMD ["/app/main"]
