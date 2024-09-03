# Go Server Project

## Overview

This project is a Go server that provides a welcoming page at the `/app` endpoint. The server runs on `localhost:8080`.

# API Specification

The API documentation for this project is stored in the [`docs` folder](./docs/) under the file name [`API-GoServerBootdev.yaml`](./docs/API-GoServerBootdev.yaml). You can view and interact with the API documentation using tools like Swagger UI, ReDoc, or any OpenAPI-compatible viewer.

## Viewing the API Documentation in Swagger UI

To view the API documentation in Swagger UI, follow these steps:

1. **Open Swagger UI in your browser.**
   
2. **Import the [`API-GoServerBootdev.yaml`](./docs/API-GoServerBootdev.yaml) file from the `docs` folder.**

3. **Explore the available endpoints and their specifications.**

These steps will allow you to interact with and test the API endpoints directly within Swagger UI.

## Getting Started

Follow the steps below to set up and run the server locally.

### Prerequisites

Ensure that you have Go installed on your system. You can download it from [Go's official website](https://golang.org/dl/).

### Installation

1. **Clone the repository** (if applicable):
   ```bash
   git clone <repository-url>
   cd <repository-directory>
   ```
2. **Remove any unnessary package**
    ```bash
    go mod tidy
    ```
3. **Start the server on localhost:8080**
    ```bash
    go run .
    ```
Or build the program and run it directly
    ```bash
    go build .
    ./go-server
    ```

This markdown includes all necessary steps and instructions for setting up, running, and accessing the API documentation for your Go server project. Let me know if you need any changes!

