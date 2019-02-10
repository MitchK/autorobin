```
❌ DANGER! This tool performs trades on your behalf. The tool is still in development and is likely to contain bugs and may not work as expected. This might result in loss of real money. 
```

# autorobin

Autorobin is a command line tool that rebalances a selection assets according to the investment weights that were given by the user. The tool has as well has backtesting capabilities in order to compare a HOLD vs REBALANCE strategy.

Compared to a HOLD strategy, rebalancing sells assets that are overallocated (for example, if one of your position raised in value) and/or buys assets that are underallocated (for example, if one of your positions lost in value). With this approach, investors can benefit from moving markets with minimal interaction.

Currently, the tool rebalances a portfolio once per run. It can be deployed as a periodic job as to your benefit.

## How it works

- Distributes cash according to weights
- Sells overallocated positions

## Limitations

- Only PortfolioVisualizer CSVs are supported as input files
- Only Robinhood is supported as a broker (but you can backtest without the broker)
- Overallocated positions are not directly re-invested for every run at the moment. This is because the tool does not wait for sell orders to complete. 

## Compile the tool

```
go build -o autorobin ./cmd/autorobin/main.go
```
## Backtest

Backtesting is not very configurable at the moment. In addition to the CSV file, you need a Tiingo token to pull last year's market data.

For every data point from Tiingo (in this case every day), a rebalancing is performed.

The backtest will output a PNG file that shows the rebalancing strategy vs. a hold strategy.

Command line usage:
```
$ autorobin help backtest
NAME:
   autorobin backtest - backtest portfolio weights

USAGE:
   autorobin backtest [command options] [arguments...]

OPTIONS:
   --tiingo.token value, -t value  Tiingo token required to fetch historic data for backtesting [$TIINGO_TOKEN]
   --output DIR, -o DIR            Backtest output directory DIR (default: current dir) [$OUTPUT]
```

## Run rebalancing on account

For this, you need a Robinhood account with either cash or existing positions. Note that this tool will not touch existing positions that are not part of your desired portfolio (i.e. the PV CSV file).

```
$ autorobin help rebalance
NAME:
   autorobin rebalance - performs a rebalance of the portfolio on your account

USAGE:
   autorobin rebalance [command options] [arguments...]

OPTIONS:
   --robinhood.username value, -u value  Robinhood login username [$ROBINHOOD_USERNAME]
   --robinhood.password value, -p value  Robinhood login username [$ROBINHOOD_PASSWORD]
   --proceed, -y                         If set to true, it disables the order placement confirmation [$PROCEED]
```

## Feedback

Leave ideas and feedback as a GitHub issue or contact me via themitch777+autorobin at gmail dot com.
