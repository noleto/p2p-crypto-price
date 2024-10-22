package cmd

import (
	"fmt"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"sync"

	"log"
	"os"
)

const (
	apiURL    = "https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest"
	topicName = "crypto-usd-price"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "p2p-crypto-price",
	Short: "A simple demo for libp2p using pubsub to emit cryptocurrencies ticker price data",
	Long:  `Tool to learn about libp2p using pubsub. `,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func initConfig() {
	viper.SetConfigName(".p2p_crypto_price")
	viper.AddConfigPath("$HOME")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("p2p_cp")

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.p2p_crypto_price.yaml)")
}

func joinTopic(err error, ctx context.Context, node host.Host) *pubsub.Topic {
	// Create a new PubSub service
	ps, err := pubsub.NewGossipSub(ctx, node)
	if err != nil {
		log.Fatalf("failed to create new PubSub service: %v", err)
	}

	// Join the topic
	topic, err := ps.Join(topicName)
	if err != nil {
		log.Fatalf("failed to join topic: %v", err)
	}
	return topic
}

func createNode(ctx context.Context) (host.Host, error) {
	// Create a new libp2p host with relay and NAT traversal capabilities
	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.EnableRelay(),      // Enable relay capabilities
		libp2p.EnableNATService(), // Enable NAT traversal capabilities
	)
	if err != nil {
		log.Fatalf("failed to create libp2p node: %v", err)
	}

	go discoverPeers(ctx, node)

	// print the node's PeerInfo in multiaddr format
	peerInfo := peer.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}
	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		panic(err)
	}
	fmt.Println("libp2p node address:", addrs[0])

	return node, err
}

func initDHT(ctx context.Context, h host.Host) *dht.IpfsDHT {
	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := dht.New(ctx, h)
	if err != nil {
		panic(err)
	}
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := h.Connect(ctx, *peerinfo); err != nil {
				fmt.Println("Bootstrap warning:", err)
			}
		}()
	}
	wg.Wait()

	return kademliaDHT
}

func discoverPeers(ctx context.Context, h host.Host) {
	kademliaDHT := initDHT(ctx, h)
	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	dutil.Advertise(ctx, routingDiscovery, topicName)

	// Look for others who have announced and attempt to connect to them
	anyConnected := false
	for !anyConnected {
		fmt.Println("Searching for peers...")
		peerChan, err := routingDiscovery.FindPeers(ctx, topicName)
		if err != nil {
			panic(err)
		}
		for p := range peerChan {
			if p.ID == h.ID() {
				continue // No self connection
			}
			err := h.Connect(ctx, p)
			if err != nil {
				//fmt.Printf("Failed connecting to %s, error: %s\n", p.ID, err)
			} else {
				fmt.Println("Connected to:", p.ID)
				anyConnected = true
			}
		}
	}
	fmt.Println("Peer discovery complete")
}
