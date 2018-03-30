package main

import (
	"log"
	"os"

	logs "github.com/appscode/go/log/golog"
	"github.com/appscode/plugin-webhook/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	cmd := &cobra.Command{
		Use: "webhook [command]",
		PersistentPreRun: func(c *cobra.Command, args []string) {
			c.Flags().VisitAll(func(flag *pflag.Flag) {
				log.Printf("FLAG: --%s=%q", flag.Name, flag.Value)
			})
		},
	}
	cmd.AddCommand(pkg.NewCmdServer())
	cmd.Execute()
	os.Exit(0)
}
