# Go URL Shortener

This is a simple URL shortener service built with Go and Redis.

## Overview

The service provides an API to shorten long URLs. It uses Redis as a storage to save the mapping between the original URL and the generated short URL.

## Prerequisites

- Go
- Redis

## Installation

1. Clone the repository:

```bash
git clone https://github.com/pouyasadri/go-url-shortener.git
```
2. Navigate to the project directory:
    
    ```bash
    cd go-url-shortener
    ```
3. Install dependencies:

    ```bash
    go mod download
    ```
## Running the Application
1. Start the Redis server.

2. Run the Go application:
   ```bash
   go run main.go
   ```
## Usage
- To create a short URL, make a POST request to `/create-short-url` with a JSON body containing `long_url` and `user_id`.

- To redirect to the original URL using the short URL, make a GET request to `/{shortUrl}`.
