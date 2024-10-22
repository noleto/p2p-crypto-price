# P2P Crypto Price Feed

A Go-based tool that demonstrates the use of LibP2P to publish and subscribe to cryptocurrency price updates. The application uses a publish-subscribe (pubsub) mechanism to emit and listen to real-time cryptocurrency prices, leveraging the LibP2P framework for peer-to-peer communication.

## Overview

This tool showcases how to create a decentralized application that shares cryptocurrency price information between peers. The main features include:

- **Produce Command**: Continuously fetch cryptocurrency price updates and broadcast them to a LibP2P topic.
- **Listen Command**: Subscribe to the topic and receive real-time cryptocurrency price updates from other nodes.

### Key Components

- **LibP2P Host**: Creates a peer-to-peer network with relay and NAT traversal capabilities.
- **PubSub Topic**: Uses LibP2P's GossipSub protocol to join and broadcast messages to a pubsub topic.
- **Commands**:
  - `produce`: Fetches cryptocurrency price data and broadcasts it.
  - `listen`: Listens for broadcasts of cryptocurrency price data from other peers.

## Installation

Ensure you have [Go](https://golang.org/doc/install) installed on your machine.

1. Clone the repository:
   ```sh
   git clone https://github.com/noleto/p2p-crypto-price.git
   cd p2p-crypto-price
   ```

2. Install dependencies:
   ```sh
   go mod tidy
   ```

3. Build the application:
   ```sh
   go build -o p2p-crypto-price
   ```

## Usage

The application provides two main commands, `produce` and `listen`, to either broadcast or subscribe to cryptocurrency prices.

### Produce Command

The `produce` command fetches cryptocurrency prices and broadcasts them to the peer-to-peer network.

#### Usage

```sh
./p2p-crypto-price produce [flags]
```

#### Flags

- `-s`, `--symbol`: Symbol of the cryptocurrency to fetch price data for (default: `BTC`).
- `-q`, `--quote-refresh`: Time in seconds to refresh and publish the price quotes (default: `30`).

#### Example

To produce price data for Bitcoin (`BTC`), updated every 30 seconds:
/!\ Create a conf file in your $HOME named: "$HOME/.p2p_crypto_price.yaml"

```yaml
apiKey: <Coinmarket API KEY>
```

```sh
./p2p-crypto-price produce -s BTC -q 30
```

### Listen Command

The `listen` command subscribes to the topic and receives price updates from other peers.

#### Usage

```sh
./p2p-crypto-price listen
```

#### Example

To listen to the price updates of Bitcoin being broadcasted on the topic:

```sh
./p2p-crypto-price listen
```

## How It Works

1. **Node Creation**: Both the producer and listener create a LibP2P node with relay and NAT traversal capabilities.
2. **Join Topic**: Nodes join a specific topic (`crypto-usd-price`) using the GossipSub protocol.
3. **Produce**: The producer fetches cryptocurrency prices and publishes messages to the topic.
4. **Listen**: Listeners receive and log messages from the topic.

The tool uses CoinMarketCap's API to retrieve real-time price data, which is then shared across the peers in the network.

## Examples

- **Start a Producer Node**: This node will fetch price updates and publish them to the network.
  ```sh
  ./p2p-crypto-price produce -s ETH -q 60
  ```
  This command fetches the price of Ethereum (`ETH`) every 60 seconds and broadcasts it.

- **Start a Listener Node**: This node will listen to updates from producer nodes.
  ```sh
  ./p2p-crypto-price listen
  ```
  This command listens for price updates of cryptocurrencies on the network.

## Contributing

Feel free to submit issues, fork the repository, and send pull requests if you have any improvements to make.

## License

This project is licensed under the MIT License.

