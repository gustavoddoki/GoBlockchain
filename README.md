# GoBlockchain

This is a simple blockchain project written in Go.

## Description

This project implements a basic blockchain with features such as wallet creation, transaction sending, and balance checking. It also provides a command-line interface (CLI) to interact with the blockchain.

## Prerequisites

Make sure you have Go installed on your machine before running this project.

## Installation

1. Clone this repository:

    ```bash
    git clone https://github.com/gustavoddoki/GoBlockchain.git
    ```

2. Navigate to the project directory:

    ```bash
    cd GoBlockchain
    ```

3. Run the `main.go` file:

    ```bash
    go run main.go
    ```

## Usage

Here are the available commands in the command-line interface (CLI):

- `getbalance`: Get the balance for a specific address.
- `createblockchain`: Create a new blockchain and send the genesis block reward to a specific address.
- `printchain`: Print the blocks in the chain.
- `send`: Send a specific amount of coins from one wallet to another.
- `createwallet`: Create a new wallet.
- `listaddresses`: List the addresses in our wallet file.

Usage example:

```bash
# Get the balance for a specific address
./GoBlockchain getbalance -address ADDRESS

# Create a new blockchain
./GoBlockchain createblockchain -address ADDRESS

# Print the blocks in the chain
./GoBlockchain printchain

# Send coins from one wallet to another
./GoBlockchain send -from FROM -to TO -amount AMOUNT

# Create a new wallet
./GoBlockchain createwallet

# List the addresses in the wallet file
./GoBlockchain listaddresses
