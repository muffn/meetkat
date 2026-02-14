FROM node:22-alpine AS css

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY web/static/css/input.css ./web/static/css/
COPY web/templates/ ./web/templates/
RUN npx @tailwindcss/cli -i web/static/css/input.css -o web/static/css/style.css --minify

FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o meetkat .

FROM alpine:3.21

WORKDIR /app

COPY --from=build /app/meetkat .
COPY --from=build /app/web/templates/ ./web/templates/
COPY --from=css /app/web/static/css/style.css ./web/static/css/
COPY --from=build /app/web/static/js/ ./web/static/js/
COPY --from=build /app/web/static/icons/ ./web/static/icons/
COPY --from=build /app/web/static/manifest.json ./web/static/

RUN mkdir -p /app/data

ENV GIN_MODE=release

EXPOSE 8080

CMD ["./meetkat"]
