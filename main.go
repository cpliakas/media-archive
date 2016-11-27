package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd handles the root "media-archive" command.
var RootCmd = &cobra.Command{
	Use:   "media-archive",
	Short: "A tool that archives personal media files",
	Long:  `A tool that archives personal media files to various storage backends, e.g. AWS S3.`,
	Run:   RunRootCmd,
}

// RunRootCmd is the work function for RootCmd.
func RunRootCmd(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(context.Background())
	EventListener(cancel)

	root := viper.GetString("root-dir")

	out1, _ := DirectoryWatcher(ctx, root)
	out2, _ := DirectoryScanner(ctx, root)
	ArchiveMedia(out1, out2)

	<-ctx.Done()
}

func main() {

	viper.SetEnvPrefix("MEDIA_ARCHIVE")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	AddSubcommands(RootCmd)
	InitGlobalConfig(RootCmd)
	InitRootCmdConfig(RootCmd)

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

// EventListener listens for shutdown signals and calls the context's cancel
// function when received.
func EventListener(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for sig := range c {
			log.Printf("shutdown signal received [signal=%s]", sig)
			cancel()
			break
		}
	}()
}

// AddSubcommands adds subcommands (usually) to the root command.
func AddSubcommands(cmd *cobra.Command) {
	//	cmd.AddCommand(SomeSubCmd)
}

// InitGlobalConfig adds global configuration options.
func InitGlobalConfig(cmd *cobra.Command) {

	cmd.PersistentFlags().BoolP("debug", "d", false, "Show debug level log messages.")
	viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))
	viper.SetDefault("debug", false)
}

// InitRootCmdConfig adds configuration options specific to RootCmd.
func InitRootCmdConfig(cmd *cobra.Command) {

	cmd.Flags().StringP("root-dir", "D", ".", "The root directory that media files are contained under.")
	viper.BindPFlag("root-dir", cmd.Flags().Lookup("root-dir"))
	viper.SetDefault("root-dir", ".")
}
