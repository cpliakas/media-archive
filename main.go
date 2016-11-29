package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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

	out1, err1 := DirectoryWatcher(ctx, root)
	out2, err2 := DirectoryScanner(ctx, root)
	err3 := ArchiveMedia(out1, out2)
	HandleErrors(err1, err2, err3)

	<-ctx.Done()
}

var ListBucketsCmd = &cobra.Command{
	Use:   "list-buckets",
	Short: "List S3 buckets",
	Long:  `List S3 buckets.`,
	Run:   RunListBucketsCmd,
}

func RunListBucketsCmd(cmd *cobra.Command, args []string) {
	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	svc := s3.New(sess)

	params := &s3.ListBucketsInput{}

	resp, err := svc.ListBuckets(params)
	if err != nil {
		log.Fatal(err)
	}

	for _, bucket := range resp.Buckets {
		fmt.Println(*bucket.Name)
	}
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
	cmd.AddCommand(ListBucketsCmd)
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
