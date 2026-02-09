FROM node:22-alpine AS css

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY static/css/input.css ./static/css/
COPY templates/ ./templates/
RUN npx @tailwindcss/cli -i static/css/input.css -o static/css/style.css --minify

FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o meetkat .

FROM alpine:3.21

WORKDIR /app

COPY --from=build /app/meetkat .
COPY --from=build /app/templates/ ./templates/
COPY --from=css /app/static/css/style.css ./static/css/

RUN mkdir -p /app/data

EXPOSE 8080

CMD ["./meetkat"]
