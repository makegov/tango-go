// list-contracts is a tiny example that prints the most recent DoD
// contracts using the Tango Go SDK.
//
// Run with: TANGO_API_KEY=… go run ./examples/list-contracts
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/makegov/tango-go"
)

func main() {
	if os.Getenv("TANGO_API_KEY") == "" {
		log.Fatal("set TANGO_API_KEY to run this example")
	}

	client := tango.NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	page, err := client.ListContracts(ctx, &tango.ListContractsOptions{
		ListOptions: tango.ListOptions{
			Limit: 10,
			Shape: tango.ShapeContractsMinimal,
		},
		AwardingAgency: "9700", // Department of Defense
		Sort:           "award_date",
		Order:          "desc",
	})
	if err != nil {
		var rle *tango.RateLimitError
		if errors.As(err, &rle) {
			log.Fatalf("rate limited (retry after %ds)", rle.RetryAfter)
		}
		log.Fatalf("list contracts: %v", err)
	}

	fmt.Printf("got %d / %d contracts\n", len(page.Results), page.Count)
	for _, c := range page.Results {
		fmt.Printf("  %s  %v  %v\n", c["award_date"], c["piid"], c["total_contract_value"])
	}

	if info := client.RateLimitInfo(); info != nil {
		fmt.Printf("\nrate-limit: %d/%d remaining (resets in %ds)\n",
			info.Remaining, info.Limit, info.ResetIn)
	}
}
