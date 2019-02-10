package rebalance

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/MitchK/autorobin/lib/autopilot"
	robinhoodBroker "github.com/MitchK/autorobin/lib/broker/robinhood"
	"github.com/MitchK/autorobin/lib/model"
)

// Run Executes rebalancing on robinhood account
func Run(desiredWeights model.Weights, assets []model.Asset, username string, password string, proceed bool) error {

	// Connect to Robinhood
	broker, err := robinhoodBroker.NewBroker(username, password)
	if err != nil {
		return err
	}
	fmt.Println("Successfully connected to Robinhood")

	// Create autopilot
	autopilot, err := autopilot.NewAutopilot(broker)
	if err != nil {
		return err
	}

	orders, err := autopilot.Rebalance(desiredWeights, false, assets...)
	if err != nil {
		return err
	}

	if len(orders) == 0 {
		fmt.Println("No orders created.")
		return nil
	}

	fmt.Println("Created the following orders:")
	for i, order := range orders {
		if order.Type == model.OrderTypeBuy {
			fmt.Printf(
				"[%d]: BUY %v x %s @ %v\n",
				i,
				order.Quantity,
				order.Asset.Symbol,
				order.Price,
			)
		} else if order.Type == model.OrderTypeSell {
			fmt.Printf(
				"[%d]: SELL %v x %s @ %v\n",
				i,
				order.Quantity,
				order.Asset.Symbol,
				order.Price,
			)
		}
	}

	if !proceed {
		proceed = askForConfirmation("Should I proceed?")
		if !proceed {
			return nil
		}
	}

	fmt.Println("Submitting orders...")
	errs := broker.Execute(orders...)
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Println(err)
		}
		return fmt.Errorf("%d/%d orders failed", len(errs), len(orders))
	}
	return nil
}

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [y/n]: ", s)
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		response = strings.ToLower(strings.TrimSpace(response))
		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
