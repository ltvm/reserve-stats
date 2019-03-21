package main

import (
	"fmt"
	"log"
	"os"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/KyberNetwork/reserve-stats/accounting/common"
	"github.com/KyberNetwork/reserve-stats/accounting/reserve-transaction-fetcher/fetcher"
	libapp "github.com/KyberNetwork/reserve-stats/lib/app"
	"github.com/KyberNetwork/reserve-stats/lib/etherscan"
)

const (
	addressesFlag = "addresses"

	fromBlockFlag = "from-block"
	toBlockFlag   = "to-block"
)

func main() {
	app := libapp.NewApp()
	app.Name = "Accounting Reserve Transaction Fetcher"
	app.Usage = "Accounting Reserve Transaction Fetcher"
	app.Action = run
	app.Version = "0.0.1"

	app.Flags = append(app.Flags,
		cli.StringSliceFlag{
			Name:   addressesFlag,
			EnvVar: "ADDRESSES",
			Usage:  "list of addresses to fetch transactions",
		},
		cli.StringFlag{
			Name:   fromBlockFlag,
			Usage:  "Fetch transactions from block",
			EnvVar: "FROM_BLOCK",
		},
		cli.StringFlag{
			Name:   toBlockFlag,
			Usage:  "Fetch transactions to block",
			EnvVar: "TO_BLOCK",
		},
	)
	app.Flags = append(app.Flags, libapp.NewPostgreSQLFlags(common.DefaultDB)...)
	app.Flags = append(app.Flags, etherscan.NewCliFlags()...)
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	if err := libapp.Validate(c); err != nil {
		return err
	}

	sugar, flusher, err := libapp.NewSugaredLogger(c)
	if err != nil {
		return err
	}

	defer flusher()

	addrs := c.StringSlice(addressesFlag)
	if len(addrs) == 0 {
		sugar.Info("no address provided")
		return nil
	}

	for _, addr := range addrs {
		if !ethereum.IsHexAddress(addr) {
			return fmt.Errorf("invalid address provided: address=%s", addr)
		}
	}

	fromBlock, err := libapp.ParseBigIntFlag(c, fromBlockFlag)
	if err != nil {
		return err
	}

	toBlock, err := libapp.ParseBigIntFlag(c, toBlockFlag)
	if err != nil {
		return err
	}

	etherscanClient, err := etherscan.NewEtherscanClientFromContext(c)
	if err != nil {
		return err
	}

	f := fetcher.NewEtherscanTransactionFetcher(sugar, etherscanClient)
	for _, addr := range addrs {
		normalTxs, err := f.NormalTx(ethereum.HexToAddress(addr), fromBlock, toBlock)
		if err != nil {
			return err
		}
		sugar.Infow("fetched normal transactions",
			"addr", addr,
			"txs", normalTxs,
		)

		internalTxs, err := f.InternalTx(ethereum.HexToAddress(addr), fromBlock, toBlock)
		if err != nil {
			return err
		}
		sugar.Infow("fetched internal transactions",
			"addr", addr,
			"txs", internalTxs,
		)

		transfers, err := f.ERC20Transfer(ethereum.HexToAddress(addr), fromBlock, toBlock)
		if err != nil {
			return err
		}
		sugar.Infow("fetched ERC20 transactions",
			"addr", addr,
			"txs", transfers,
		)
	}
	return nil
}