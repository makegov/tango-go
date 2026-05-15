// iterate-entities walks every entity matching a NAICS filter, printing
// the UEI and legal business name. Demonstrates the Iterator pattern.
//
// Run with: TANGO_API_KEY=… go run ./examples/iterate-entities
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/makegov/tango-go"
)

func main() {
	if os.Getenv("TANGO_API_KEY") == "" {
		log.Fatal("set TANGO_API_KEY to run this example")
	}

	client := tango.NewClient()

	iter := client.IterateEntities(context.Background(), &tango.ListEntitiesOptions{
		ListOptions: tango.ListOptions{
			Limit: 100,
			Shape: tango.ShapeEntitiesMinimal,
		},
		NAICS: "541512", // Computer systems design services
		State: "VA",
	})

	count := 0
	for iter.Next() {
		e := iter.Item()
		fmt.Printf("%-12s  %s\n", e["uei"], e["legal_business_name"])
		count++
		if count >= 25 {
			// Stop early for the example; iter would happily walk the
			// whole result set if we kept going.
			break
		}
	}
	if err := iter.Err(); err != nil {
		log.Fatalf("iterate: %v", err)
	}
	fmt.Printf("\nprinted %d entities\n", count)
}
