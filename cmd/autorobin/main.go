package main

import (
	"bufio"
	"errors"
	"fmt"

	"github.com/MitchK/autorobin/lib/portfolioparser/portfoliovisualizer"

	"os"

	"github.com/MitchK/autorobin/cmd/autorobin/backtest"
	"github.com/MitchK/autorobin/cmd/autorobin/rebalance"
	"github.com/MitchK/autorobin/lib/model"

	"github.com/urfave/cli"
)

const (
	date = iota
	close
	volume
	open
	high
	low
)

func main() {
	app := cli.NewApp()
	app.Name = "autorobin"
	app.Usage = "Backtests and executes portfolio relancing"

	// portfolio file
	var pvCSVfile string
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "pv.csvfile, f",
			Value:       "",
			Usage:       "Load desired portfolio weights from `FILE` (currently, portfoliovisualizer.com only)",
			Destination: &pvCSVfile,
		},
	}

	// backtest command
	var tiingoToken string
	var output string
	backtest := cli.Command{
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "tiingo.token, t",
				Value:       "",
				EnvVar:      "TIINGO_TOKEN",
				Usage:       "Tiingo token required to fetch historic data for backtesting",
				Destination: &tiingoToken,
			},
			cli.StringFlag{
				Name:        "output, o",
				EnvVar:      "OUTPUT",
				Value:       "",
				Usage:       "Backtest output directory `DIR` (default: current dir)",
				Destination: &output,
			},
		},
		Name:    "backtest",
		Aliases: []string{"b"},
		Usage:   "backtest portfolio weights",
		Action: func(c *cli.Context) error {
			if pvCSVfile == "" {
				return errors.New("No PortfolioVisualizer csv file provided")
			}
			if tiingoToken == "" {
				return errors.New("No Tiingo token provided")
			}
			weights, assets, err := parsePVFile(pvCSVfile)
			if err != nil {
				return nil
			}
			return backtest.Run(weights, assets, tiingoToken, output)
		},
	}

	// rebalance command
	var robinhoodUsername string
	var robinhoodPassword string
	var proceed bool
	rebalance := cli.Command{
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "robinhood.username, u",
				EnvVar:      "ROBINHOOD_USERNAME",
				Value:       "",
				Usage:       "Robinhood login username",
				Destination: &robinhoodUsername,
			},
			cli.StringFlag{
				Name:        "robinhood.password, p",
				EnvVar:      "ROBINHOOD_PASSWORD",
				Value:       "",
				Usage:       "Robinhood login username",
				Destination: &robinhoodPassword,
			},
			cli.BoolFlag{
				Name:        "proceed, y",
				EnvVar:      "PROCEED",
				Usage:       "If set to true, it disables the order placement confirmation",
				Destination: &proceed,
			},
		},
		Name:    "rebalance",
		Aliases: []string{"r"},
		Usage:   "performs a rebalance of the portfolio on your account",
		Action: func(c *cli.Context) error {
			if pvCSVfile == "" {
				return errors.New("No PortfolioVisualizer csv file provided")
			}
			weights, assets, err := parsePVFile(pvCSVfile)
			if err != nil {
				return err
			}
			if robinhoodUsername == "" {
				return errors.New("No Robinhood username provided")
			}
			if robinhoodPassword == "" {
				return errors.New("No Robinhood password provided")
			}
			return rebalance.Run(weights, assets, robinhoodUsername, robinhoodPassword, proceed)
		},
	}

	app.Commands = []cli.Command{
		backtest,
		rebalance,
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func parsePVFile(path string) (model.Weights, []model.Asset, error) {
	csvFile, err := os.Open(path)
	if err != nil {
		return model.Weights{}, []model.Asset{}, err
	}
	parser := portfoliovisualizer.NewParser()
	return parser.Parse(bufio.NewReader(csvFile))
}
