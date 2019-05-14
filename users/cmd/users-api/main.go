package main

import (
	"log"
	"os"

	"github.com/KyberNetwork/tokenrate/coingecko"
	"github.com/urfave/cli"

	libapp "github.com/KyberNetwork/reserve-stats/lib/app"
	"github.com/KyberNetwork/reserve-stats/lib/httputil"
	libredis "github.com/KyberNetwork/reserve-stats/lib/redis"
	usercommon "github.com/KyberNetwork/reserve-stats/users/common"
	"github.com/KyberNetwork/reserve-stats/users/http"
)

func main() {
	app := libapp.NewApp()
	app.Name = "User stat module"
	app.Usage = "Store and return user stat information"
	app.Action = run
	app.Version = "0.0.1"

	app.Flags = append(app.Flags, usercommon.NewUserCapCliFlags()...)
	app.Flags = append(app.Flags, httputil.NewHTTPCliFlags(httputil.UsersPort)...)
	app.Flags = append(app.Flags, libredis.NewCliFlags()...)
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	if err := libapp.Validate(c); err != nil {
		return err
	}

	sugar, flush, err := libapp.NewSugaredLogger(c)
	if err != nil {
		return err
	}
	defer flush()

	sugar.Info("Run user module")

	redisCacheClient, err := libredis.NewClientFromContext(c)
	if err != nil {
		return err
	}
	userCapConf := usercommon.NewUserCapConfigurationFromContext(c)

	server := http.NewServer(
		sugar,
		coingecko.New(),
		httputil.NewHTTPAddressFromContext(c),
		redisCacheClient,
		userCapConf)
	return server.Run()
}
