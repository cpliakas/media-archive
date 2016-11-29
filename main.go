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

	archive := viper.GetString("archive-name")
	bucket := viper.GetString("aws-bucket")
	root := viper.GetString("root-dir")

	watcher, err := NewMediaWatcher(root)
	if err != nil {
		panic(err)
	}

	err2 := ArchiveMedia(watcher, archive, bucket)
	HandleErrors(watcher.Errors(), err2)

	<-ctx.Done()
}

var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "For testing shit",
	Long:  `For testing shit.`,
	Run:   RunTestCmd,
}

func RunTestCmd(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(context.Background())
	EventListener(cancel)

	root := viper.GetString("root-dir")
	w, err := NewMediaWatcher(root)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	go func() {
		for {
			select {
			case path := <-w.Media():
				log.Println("INFO discovered file:", w.RelativePath(path))
			case err := <-w.Errors():
				log.Println("ERROR", err)
			}
		}
	}()

	<-ctx.Done()
}

func main() {

	viper.SetEnvPrefix("MEDIA_ARCHIVE")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	AddSubcommands(RootCmd)
	InitGlobalConfig(RootCmd)
	InitRootCmdConfig(RootCmd)
	InitTestCmdConfig(TestCmd)

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
	cmd.AddCommand(TestCmd)
}

// InitGlobalConfig adds global configuration options.
func InitGlobalConfig(cmd *cobra.Command) {

	cmd.PersistentFlags().BoolP("debug", "d", false, "Show debug level log messages.")
	viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))
	viper.SetDefault("debug", false)

	cmd.Flags().StringP("archive-name", "n", "media-archive", "The name of the archive, e.g. my-photos.")
	viper.BindPFlag("archive-name", cmd.Flags().Lookup("archive-name"))
	viper.SetDefault("archive-name", "media-archive")
}

// InitRootCmdConfig adds configuration options specific to RootCmd.
func InitRootCmdConfig(cmd *cobra.Command) {

	cmd.Flags().StringP("root-dir", "D", ".", "The root directory that media files are contained under.")
	viper.BindPFlag("root-dir", cmd.Flags().Lookup("root-dir"))
	viper.SetDefault("root-dir", ".")

	cmd.Flags().StringP("aws-bucket", "b", "", "The AWS S3 bucket that media files are archived to.")
	viper.BindPFlag("aws-bucket", cmd.Flags().Lookup("aws-bucket"))
	viper.SetDefault("aws-bucket", "")
}

// InitTestCmdConfig adds configuration options specific to TestCmd.
func InitTestCmdConfig(cmd *cobra.Command) {

	cmd.Flags().StringP("root-dir", "D", ".", "The root directory that media files are contained under.")
	viper.BindPFlag("root-dir", cmd.Flags().Lookup("root-dir"))
	viper.SetDefault("root-dir", ".")
}
