package main

import (
	"log"
	"os"
	"time"

	libapp "github.com/KyberNetwork/reserve-stats/lib/app"
	"github.com/KyberNetwork/reserve-stats/lib/influxdb"
	"github.com/KyberNetwork/reserve-stats/lib/timeutil"
	tokenrate "github.com/KyberNetwork/reserve-stats/tokenratefetcher"
	"github.com/KyberNetwork/reserve-stats/tokenratefetcher/storage"
	"github.com/KyberNetwork/reserve-stats/tokenratefetcher/storage/postgres"
	"github.com/KyberNetwork/tokenrate/coingecko"
	"github.com/urfave/cli"
)

const (
	defaultFromTime     = "2018-01-01T00:00:00Z"
	kyberNetworkTokenID = "kyber-network"
	usdCurrencyID       = "usd"
	dbName              = "token_rate"

	dbEngineFlag    = "db-engine"
	defaultDBEngine = "postgres"
)

func main() {
	app := libapp.NewApp()
	app.Name = "KNC USD rate Fetcher"
	app.Usage = "Fetch KNC-USD Rate from provider"
	app.Version = "0.0.1"
	app.Action = run
	app.Flags = append(app.Flags, timeutil.NewTimeRangeCliFlags()...)
	app.Flags = append(app.Flags, influxdb.NewCliFlags()...)
	app.Flags = append(app.Flags, libapp.NewPostgreSQLFlags(dbName)...)
	app.Flags = append(app.Flags,
		cli.StringFlag{
			Name:   dbEngineFlag,
			Usage:  "engine database (influxdb or postgres)",
			EnvVar: "DB_ENGINE",
			Value:  defaultDBEngine, // default value
		},
	)
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	var (
		dbStorage storage.Interface
	)
	sugar, flush, err := libapp.NewSugaredLogger(c)
	if err != nil {
		return err
	}
	defer flush()

	cgk := coingecko.New()

	if c.String(dbEngineFlag) != "postgres" {
		influxClient, err := influxdb.NewClientFromContext(c)
		if err != nil {
			return err
		}

		dbStorage, err = storage.NewInfluxStorage(influxClient, dbName, sugar)
		if err != nil {
			return err
		}
	} else {
		db, err := libapp.NewDBFromContext(c)
		if err != nil {
			return err
		}
		if dbStorage, err = postgres.NewPostgresStorage(sugar, db); err != nil {
			return err
		}
	}

	tokenRate, err := tokenrate.NewRateFetcher(sugar, dbStorage, cgk)
	if err != nil {
		return err
	}

	from, err := timeutil.FromTimeFromContext(c)
	if err == timeutil.ErrEmptyFlag {
		sugar.Debug("no from time provided, seeking for the first data point in DB...")
		from, err = dbStorage.LastTimePoint(cgk.Name(), kyberNetworkTokenID, usdCurrencyID)
		if err != nil {
			return err
		}

		if from.IsZero() {
			if from, err = time.Parse(time.RFC3339, defaultFromTime); err != nil {
				return err
			}
			sugar.Infow("no record found in database, using default from time",
				"from", from,
			)
		} else {
			sugar.Infow("found last timestamp in database",
				"from", from,
			)
		}

		// starts with the day after the day stored in database
		from = from.AddDate(0, 0, 1)
	} else if err != nil {
		return err
	}

	to, err := timeutil.ToTimeFromContext(c)
	if err == timeutil.ErrEmptyFlag {
		to = time.Now().UTC()
		sugar.Info("no to time provided, using current timestamp",
			"to", to)
	} else if err != nil {
		return err
	}

	return tokenRate.FetchRatesInRanges(from, to, kyberNetworkTokenID, usdCurrencyID)
}
