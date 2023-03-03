# BattleBlocks Graffle Webhook Server

This repository contains the webhook server for BattleBlocks. The webhook server is responsible for issuing event notifications to a Google Cloud Pub/Sub topic and triggering updates to the game state.

## Project Structure

The repository has the following structure:

    ├── Dockerfile
    ├── README.md
    ├── cmd
    │   └── main.go
    ├── go.mod
    ├── go.sum
    ├── pubsub
    │   └── pubsub_client.go
    └── webhook
        └── webhook_server.go

- `Dockerfile`: contains instructions to build a Docker image for the Graffle Webhook.
- `README.md`: this file, which provides an overview of the repository.
- `cmd/main.go`: the main entry point for the application, which initializes the server and starts listening for incoming requests.
- `go.mod` and `go.sum`: the Go module files, which manage the project's dependencies.
- `pubsub/pubsub_client.go`: a package that contains a client for communicating with the Google Cloud Pub/Sub service.
- `webhook/webhook_server.go`: a package that contains the main logic for the Graffle Webhook.

## Usage

To use the Graffle Webhook, you must first set up a Google Cloud Pub/Sub topic and subscription. Once you have those set up, you can deploy the Graffle Webhook to a server or run it locally.

### Deploying to a Server

To deploy the Graffle Webhook to a server, follow these steps:

1. Build the Docker image: `docker build -t graffle-webhook .`
2. Tag the image: `docker tag graffle-webhook <your-registry>/graffle-webhook:<version>`
3. Push the image to your container registry: `docker push <your-registry>/graffle-webhook:<version>`
4. Deploy the container to your server using a tool like Kubernetes or Docker Compose.

### Running Locally

To run the Graffle Webhook locally, you can use the `go run` command:

- `go run cmd/main.go`

## Dependencies

The Graffle Webhook depends on several Go packages, which are managed using Go modules. The dependencies are listed in the `go.mod` file.
