package cmd

import (
	"encoding/json"
	"fmt"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"io"
	"log"
	"net/http"
	"time"
)

type CryptoQuoteResponse struct {
	Data map[string][]struct {
		Quote map[string]struct {
			Price float64 `json:"price"`
		} `json:"quote"`
	} `json:"data"`
}

var (
	cryptoSymbol     string
	quoteRefreshSecs int
)

func getCryptoPrice(symbol string) (float64, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", viper.GetString("apiKey"))

	q := req.URL.Query()
	q.Add("symbol", symbol)
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	var response CryptoQuoteResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}

	if cryptoData, exists := response.Data[symbol]; exists && len(cryptoData) > 0 {
		price := cryptoData[0].Quote["USD"].Price
		return price, nil
	}

	return 0, fmt.Errorf("could not find price for symbol %s", symbol)
}

func publishMessage(ctx context.Context, topic *pubsub.Topic, message string) error {
	return topic.Publish(ctx, []byte(message))
}

func produceRun(cmd *cobra.Command, args []string) {
	fmt.Printf("Starting producer for ticker %v every %v seconds...\n", cryptoSymbol, quoteRefreshSecs)
	ctx := context.Background()

	node, err := createNode(ctx)
	defer node.Close()

	topic := joinTopic(err, ctx, node)

	emitCryptoFeed(ctx, topic)
}

func emitCryptoFeed(ctx context.Context, topic *pubsub.Topic) {
	ticker := time.NewTicker(time.Duration(quoteRefreshSecs) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			price, err := getCryptoPrice(cryptoSymbol)
			if err != nil {
				log.Printf("failed to get price: %v", err)
				continue
			}

			message := fmt.Sprintf("%s: $%.2f", cryptoSymbol, price)
			if err := publishMessage(ctx, topic, message); err != nil {
				log.Printf("failed to publish message: %v", err)
			}
			log.Printf("Emitting message: %s\n", message)
		}
	}
}

// produceCmd represents the produce command
var produceCmd = &cobra.Command{
	Use:   "produce",
	Short: "Produce crypto price feed",
	Long:  `Produce crypto price feed`,
	Run:   produceRun,
}

func init() {
	rootCmd.AddCommand(produceCmd)

	produceCmd.Flags().StringVarP(&cryptoSymbol, "symbol", "s", "BTC", "Crypto price symbol")
	produceCmd.Flags().IntVarP(&quoteRefreshSecs, "quote-refresh", "q", 30, "Time in secs to refresh quotes")
}
