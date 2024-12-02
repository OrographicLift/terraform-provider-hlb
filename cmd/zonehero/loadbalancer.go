package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"gitlab.guerraz.net/HLB/hlb-terraform-provider/hlb"
)

func init() {
	// Load Balancer Commands
	hlbCmd.AddCommand(listLoadBalancersCmd)
	hlbCmd.AddCommand(createLoadBalancerCmd)
	hlbCmd.AddCommand(deleteLoadBalancerCmd)
	hlbCmd.AddCommand(updateLoadBalancerCmd)
	hlbCmd.AddCommand(getLoadBalancerCmd)

	// List Load Balancers Flags
	listLoadBalancersCmd.Flags().Int("limit", 20, "Maximum number of items to return")
	listLoadBalancersCmd.Flags().String("next-token", "", "Token for pagination")

	// Create Load Balancer Flags
	createLoadBalancerCmd.Flags().StringP("name", "n", "", "Name of the load balancer")
	createLoadBalancerCmd.Flags().BoolP("internal", "i", false, "Whether the load balancer is internal")
	createLoadBalancerCmd.Flags().StringSliceP("subnets", "s", []string{}, "Subnets for the load balancer")
	createLoadBalancerCmd.Flags().StringSliceP("security-groups", "g", []string{}, "Security groups for the load balancer")
	createLoadBalancerCmd.Flags().String("input-json", "", "JSON file containing load balancer configuration")
	createLoadBalancerCmd.MarkFlagRequired("name")
	createLoadBalancerCmd.MarkFlagRequired("subnets")

	// Update Load Balancer Flags
	updateLoadBalancerCmd.Flags().String("id", "", "ID of the load balancer to update")
	updateLoadBalancerCmd.Flags().String("name", "", "New name for the load balancer")
	updateLoadBalancerCmd.Flags().String("input-json", "", "JSON file containing update configuration")
	updateLoadBalancerCmd.MarkFlagRequired("id")

	// Get Load Balancer Flags
	getLoadBalancerCmd.Flags().String("id", "", "ID of the load balancer to get")
	getLoadBalancerCmd.MarkFlagRequired("id")

	// Delete Load Balancer Flags
	deleteLoadBalancerCmd.Flags().String("id", "", "ID of the load balancer to delete")
	deleteLoadBalancerCmd.MarkFlagRequired("id")
}

var listLoadBalancersCmd = &cobra.Command{
	Use:   "list-load-balancers",
	Short: "List all load balancers",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := createClient(cmd.Context())
		if err != nil {
			return err
		}

		limit, _ := cmd.Flags().GetInt("limit")
		nextToken, _ := cmd.Flags().GetString("next-token")

		loadBalancers, newNextToken, err := client.ListLoadBalancers(cmd.Context(), limit, nextToken)
		if err != nil {
			return err
		}

		if output == "json" {
			return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
				"items":     loadBalancers,
				"nextToken": newNextToken,
			})
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tDNS NAME\tSTATE")
		for _, lb := range loadBalancers {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", lb.ID, lb.Name, lb.DNSName, lb.State)
		}
		w.Flush()

		if newNextToken != "" {
			fmt.Printf("\nUse --next-token %s to get the next page\n", newNextToken)
		}

		return nil
	},
}

var getLoadBalancerCmd = &cobra.Command{
	Use:   "get-load-balancer",
	Short: "Get details of a specific load balancer",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := createClient(cmd.Context())
		if err != nil {
			return err
		}

		id, _ := cmd.Flags().GetString("id")
		lb, err := client.GetLoadBalancer(cmd.Context(), id)
		if err != nil {
			return err
		}

		if output == "json" {
			return json.NewEncoder(os.Stdout).Encode(lb)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tDNS NAME\tSTATE")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", lb.ID, lb.Name, lb.DNSName, lb.State)
		return w.Flush()
	},
}

var createLoadBalancerCmd = &cobra.Command{
	Use:   "create-load-balancer",
	Short: "Create a new load balancer",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := createClient(cmd.Context())
		if err != nil {
			return err
		}

		var input hlb.LoadBalancerCreate

		// Check if JSON input file is provided
		if inputJSON, _ := cmd.Flags().GetString("input-json"); inputJSON != "" {
			data, err := os.ReadFile(inputJSON)
			if err != nil {
				return fmt.Errorf("failed to read input JSON file: %w", err)
			}
			if err := json.Unmarshal(data, &input); err != nil {
				return fmt.Errorf("failed to parse input JSON: %w", err)
			}
		} else {
			// Get values from flags
			name, _ := cmd.Flags().GetString("name")
			internal, _ := cmd.Flags().GetBool("internal")
			subnets, _ := cmd.Flags().GetStringSlice("subnets")
			securityGroups, _ := cmd.Flags().GetStringSlice("security-groups")

			input = hlb.LoadBalancerCreate{
				Name:           name,
				Internal:       internal,
				Subnets:        subnets,
				SecurityGroups: securityGroups,
			}
		}

		lb, err := client.CreateLoadBalancer(cmd.Context(), &input)
		if err != nil {
			return err
		}

		if output == "json" {
			return json.NewEncoder(os.Stdout).Encode(lb)
		}

		fmt.Printf("Created load balancer: %s (ID: %s)\n", lb.Name, lb.ID)
		return nil
	},
}

var updateLoadBalancerCmd = &cobra.Command{
	Use:   "update-load-balancer",
	Short: "Update an existing load balancer",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := createClient(cmd.Context())
		if err != nil {
			return err
		}

		id, _ := cmd.Flags().GetString("id")
		var input hlb.LoadBalancerUpdate

		// Check if JSON input file is provided
		if inputJSON, _ := cmd.Flags().GetString("input-json"); inputJSON != "" {
			data, err := os.ReadFile(inputJSON)
			if err != nil {
				return fmt.Errorf("failed to read input JSON file: %w", err)
			}
			if err := json.Unmarshal(data, &input); err != nil {
				return fmt.Errorf("failed to parse input JSON: %w", err)
			}
		} else {
			// Get values from flags
			if name, _ := cmd.Flags().GetString("name"); name != "" {
				input.Name = &name
			}
		}

		lb, err := client.UpdateLoadBalancer(cmd.Context(), id, &input)
		if err != nil {
			return err
		}

		if output == "json" {
			return json.NewEncoder(os.Stdout).Encode(lb)
		}

		fmt.Printf("Updated load balancer: %s (ID: %s)\n", lb.Name, lb.ID)
		return nil
	},
}

var deleteLoadBalancerCmd = &cobra.Command{
	Use:   "delete-load-balancer",
	Short: "Delete a load balancer",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := createClient(cmd.Context())
		if err != nil {
			return err
		}

		id, _ := cmd.Flags().GetString("id")
		if err := client.DeleteLoadBalancer(cmd.Context(), id); err != nil {
			return err
		}

		if output == "json" {
			fmt.Println("{\"status\": \"deleted\"}")
		} else {
			fmt.Printf("Deleted load balancer: %s\n", id)
		}

		return nil
	},
}
