FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o gowebapp .

FROM alpine:3.21

WORKDIR /app

COPY --from=build /app/gowebapp .
COPY --from=build /app/templates/ ./templates/

EXPOSE 8080

CMD ["./gowebapp"]
