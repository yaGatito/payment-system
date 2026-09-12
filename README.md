# Payment System

## Overview
This project is a Go-based payment backend made of three services:
- Wallet Service: GraphQL API for cards and payments;
- Callback Service: HTTP endpoint for provider callbacks;
- Payment Worker: NATS JetStream consumer that handles `payment`/`add-card` events.

The goal is to show proper layered architecture, async processing, callback handling, and DB consistency patterns without building a full production payment platform.

## How to set up & run

### 1. Ngrok

I used ngrok as a free-tier tunnel provider, so you need to register at https://ngrok.com/ and copy the provided token. <br>

* ```sudo apt update && sudo apt install ngrok``` -- update and install the command-line tool <br>

* ```ngrok config add-authtoken <your_token>``` -- provide the token copied after registration <br>

* ```ngrok http 8081``` -- open a tunnel for a specific port (the port should be the same as the `CALLBACK_PORT` env). Copy the provided URL and set it as the `ROZETKA_CALLBACK_URL` env <br>

### 2. Docker Compose

* ```. .example.env``` -- example of env. Here you should set the `ROZETKA_CALLBACK_URL` env to `<url-provided-by-ngrok>/notify` <br>

* ```cd ~/workdir && docker compose up```

* Proceed to http://localhost:8082/graphql (change the port if the `GRAPHQL_PORT` env was changed)

## Test cards provided by Rozetka
* `4242 4242 4242 4242`	(3D Secure) -- success
* `4111 1111 1111 1111`	(2D Secure) -- success
* `4200 0000 0000 0000` -- failed
<br> More test data is available at https://docs.rozetkapay.com/sandbox/test-cards/
## GraphQL queries

* `AddCardToWallet`: proceed using the redirect URL and provide a card, any future expiration date, and any CVV.
```
mutation {
  addCardToWallet(customerId: "e2118114-957a-48d6-b1ae-d487cdc99535") {
    redirectUrl
  }
}
```

* `GetCards`: provide a customer ID to retrieve related cards
```
query {
  getCards(customerId: "e2118114-957a-48d6-b1ae-d487cdc99535") {
    id
    last4
    linkedAt
    type
  }
}
```

* `PayWithCard`: provide a customer ID, card ID (from the `GetCards` result), amount, currency, order ID, and follow the redirect URL if 3D Secure is enabled to confirm the payment.
```
mutation {
  payWithCard(
    customerId: "e2118114-957a-48d6-b1ae-d487cdc99535"
    cardId: "b6858230-3439-4068-92f0-fb3f6567e1f4"
    amount: 120
    currency: "UAH"
    orderId: "1d955abe-9522-33d5-a788-a1b186b163d3"
  ) {
    paymentId
    status
    redirectUrl
  }
}
```

* `GetPayments`: get payment history by provided customer ID, offset, and limit.
```
query {
  getPayments(customerId: "e2118114-957a-48d6-b1ae-d487cdc99535", limit: 10,  offset: 0) {
    paymentId
    amount
    currency
    status
    createdAt
  }
}
```


## Architecture

```text
Client
  │
  ▼
Wallet GraphQL API
  │   ├── card linking
  │   ├── list cards
  │   └── charge saved card
  │
  ├── DB (PostgreSQL)
  └── RozetkaPay API provider
          │
          ▼
    Callback Service
          │
          ▼
      NATS JetStream
          │
          ▼
      Payment Worker
          │
          ▼
          DB
```

## Environment variables

Key variables used by the project include:

```bash
# Postgres config
WALLET_DB_USER
WALLET_DB_PASS
WALLET_DB_NAME
WALLET_DB_HOST
WALLET_DB_PORT

# Rozetka integration
ROZETKA_BASE_URL
ROZETKA_USERNAME
ROZETKA_PASSWORD
ROZETKA_CALLBACK_URL

# Wallet GraphQL
GRAPHQL_PORT

# Callback service
CALLBACK_PORT
NATS_PORT
NATS_MGMT_PORT
NATS_URL
NATS_STREAM
NATS_SUBJECT
```
