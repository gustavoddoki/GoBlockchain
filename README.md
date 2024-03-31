# GoBlockchain

This is a simple blockchain project written in Go.

## Description

This project implements a basic blockchain with features such as wallet creation, transaction sending, and balance checking. It also provides a command-line interface (CLI) to interact with the blockchain.

## Prerequisites

Make sure you have Go installed on your machine before running this project.

## Installation
```
git clone https://github.com/gustavoddoki/GoBlockchain.git
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

- Get the balance for a specific address
```
go run main.go getbalance -address ADDRESS
```
- Create a new blockchain
```
go run main.go createblockchain -address ADDRESS
```
- Print the blocks in the chain
```
go run main.go printchain
```
- Send coins from one wallet to another
```
go run main.go send -from FROM -to TO -amount AMOUNT
```
- Create a new wallet
```
go run main.go createwallet
```
- List the addresses in the wallet file
```
go run main.go listaddresses
```
