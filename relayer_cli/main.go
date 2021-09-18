package main

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"os"
	"path/filepath"
	"sort"
)

var (
	// The app that holds all commands and flags.
	app *cli.App

	KeyFlag = cli.StringFlag{
		Name:  "key",
		Usage: "Private key file path",
		Value: "",
	}
	KeyStoreFlag = cli.StringFlag{
		Name:  "keystore",
		Usage: "Keystore file path",
	}
	RPCListenAddrFlag = cli.StringFlag{
		Name:  "rpcaddr",
		Usage: "HTTP-RPC server listening interface",
		Value: "localhost",
	}
	RPCPortFlag = cli.IntFlag{
		Name:  "rpcport",
		Usage: "HTTP-RPC server listening port",
		Value: 7445,
	}

	FeeFlag = cli.Uint64Flag{
		Name:  "fee",
		Usage: "work fee",
		Value: 0,
	}

	PublicAdressFlag = cli.StringFlag{
		Name:  "PkAdress",
		Usage: "Relayer bft key for BFT (no 0x prefix)",
		Value: "",
	}
)

func init() {
	app = cli.NewApp()
	app.Usage = "Atlas Register Tool"
	app.Name = filepath.Base(os.Args[0])
	app.Version = "1.0.0"
	app.Copyright = "Copyright 2020-2021 The Atlas Authors"
	app.Flags = []cli.Flag{
		KeyFlag,
		KeyStoreFlag,
		RPCListenAddrFlag,
		RPCPortFlag,
		FeeFlag,
		PublicAdressFlag,
	}
	app.Action = MigrateFlags(start)
	app.CommandNotFound = func(ctx *cli.Context, cmd string) {
		fmt.Fprintf(os.Stderr, "No such command: %s\n", cmd)
		os.Exit(1)
	}
	// Add subcommands.
	app.Commands = []cli.Command{}
	cli.CommandHelpTemplate = OriginCommandHelpTemplate
	sort.Sort(cli.CommandsByName(app.Commands))
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var OriginCommandHelpTemplate string = `{{.Name}}{{if .Subcommands}} command{{end}}{{if .Flags}} [command options]{{end}} [arguments...] {{if .Description}}{{.Description}} {{end}}{{if .Subcommands}} SUBCOMMANDS:     {{range .Subcommands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}     {{end}}{{end}}{{if .Flags}} OPTIONS: {{range $.Flags}}{{"\t"}}{{.}} {{end}} {{end}}`

func MigrateFlags(action func(ctx *cli.Context) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		for _, name := range ctx.FlagNames() {
			if ctx.IsSet(name) {
				ctx.GlobalSet(name, ctx.String(name))
			}
		}
		return action(ctx)
	}
}

func start(ctx *cli.Context) error {
	commpassInfo := commpassInfo{}
	commpassInfo.relayerData = []*relayerInfo{
		{url: keystore1},
	}
	commpassInfo.preWork(ctx)
	go commpassInfo.atlasBackend()
	go commpassInfo.saveMock()
	go commpassInfo.doTxVerity()
	select {}
	return nil
}
