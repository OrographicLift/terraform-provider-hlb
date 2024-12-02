package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/spf13/cobra"
	"gitlab.guerraz.net/HLB/hlb-terraform-provider/hlb"
)

var (
	profile string
	region  string
	apiKey  string
	output  string
	debug   bool
)

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "AWS profile to use")
	rootCmd.PersistentFlags().StringVar(&region, "region", "", "AWS region to use")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "HLB API key")
	rootCmd.PersistentFlags().StringVar(&output, "output", "text", "Output format (json/text)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug output")
}

var rootCmd = &cobra.Command{
	Use:   "zonehero",
	Short: "ZoneHero CLI - Manage HLB resources",
	Long: `ZoneHero CLI provides a command-line interface to manage HLB (Hero Load Balancer) resources.
It supports managing load balancers and listeners with various operations like create, list, update, and delete.`,
}

var hlbCmd = &cobra.Command{
	Use:   "hlb",
	Short: "Manage HLB resources",
	Long:  `Manage HLB (Hero Load Balancer) resources including load balancers and listeners.`,
}

func createClient(ctx context.Context) (*hlb.Client, error) {
	if apiKey == "" {
		apiKey = os.Getenv("HLB_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("HLB API key is required. Set it using --api-key flag or HLB_API_KEY environment variable")
	}

	opts := []func(*config.LoadOptions) error{}
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("error loading AWS config: %v", err)
	}

	client, err := hlb.NewClient(ctx, apiKey, awsCfg)
	if err != nil {
		return nil, fmt.Errorf("error creating HLB client: %v", err)
	}

	client.SetDebug(debug)
	return client, nil
}

func main() {
	rootCmd.AddCommand(hlbCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
