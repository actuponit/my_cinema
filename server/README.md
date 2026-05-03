# Go Server

This folder contains the Go API server for the cinema app.

## Requirements

- Go 1.22.5 or newer
- A reachable Hasura GraphQL endpoint
- A valid Cloudinary account if you want image uploads to work
- Chapa credentials if you want payment initialization and webhook handling to work

## Environment Variables

The server reads these variables at runtime:

Required for startup and core API access:

- `PORT` - HTTP port the server listens on, for example `3000`
- `HASURA_GRAPHQL_ENDPOINT` - Hasura GraphQL endpoint used by the Go client
- `HASURA_GRAPHQL_ADMIN_SECRET` - admin secret sent on Hasura requests
- `JWT_SECRET` - secret used to sign and validate JWTs

Required for payment features:

- `CHAPA_SECRET_KEY` - Chapa API secret key
- `CHAPA_WEBHOOK_SECRET` - Chapa webhook HMAC secret

Required for image uploads:

- `CLOUDINARY_URL` - Cloudinary connection string used by the upload middleware

Optional MQTT settings:

- `MQTT_BROKER_URL` - defaults to `tcp://test.mosquitto.org:1883`
- `MQTT_CLIENT_ID` - defaults to `cinema-server`
- `MQTT_USERNAME` - optional broker username
- `MQTT_PASSWORD` - optional broker password
- `MQTT_TIMEOUT_SECONDS` - defaults to `30`

## Setup

1. Copy the example env file from the repository root:

```bash
cp ../.env.example ../.env
```

2. Fill in the real values in `../.env`.

3. Load the variables into your shell before running the server:

```bash
set -a
source ../.env
set +a
```

## Run Locally

From this `server` directory:

```bash
go mod download
go run ./delivery
```

The server listens on `http://localhost:$PORT`.

## Run With Docker Compose

If you prefer Docker, run the stack from the repository root after creating `.env` there:

```bash
docker compose up --build
```
