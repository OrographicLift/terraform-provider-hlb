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
	// Listener Commands
	hlbCmd.AddCommand(listListenersCmd)
	hlbCmd.AddCommand(createListenerCmd)
	hlbCmd.AddCommand(deleteListenerCmd)
	hlbCmd.AddCommand(updateListenerCmd)
	hlbCmd.AddCommand(getListenerCmd)

	// List Listeners Flags
	listListenersCmd.Flags().String("load-balancer-id", "", "ID of the load balancer")
	listListenersCmd.MarkFlagRequired("load-balancer-id")

	// Create Listener Flags
	createListenerCmd.Flags().String("load-balancer-id", "", "ID of the load balancer")
	createListenerCmd.Flags().Int("port", 0, "Port number")
	createListenerCmd.Flags().String("protocol", "", "Protocol (HTTP/HTTPS)")
	createListenerCmd.Flags().String("target-group-arn", "", "Target group ARN")
	createListenerCmd.Flags().String("certificate-secrets-name", "", "Certificate secrets name (for HTTPS)")
	createListenerCmd.Flags().String("alpn-policy", "", "ALPN policy")
	createListenerCmd.Flags().Bool("enable-deletion-protection", false, "Enable deletion protection")
	createListenerCmd.Flags().String("input-json", "", "JSON file containing listener configuration")
	createListenerCmd.MarkFlagRequired("load-balancer-id")
	createListenerCmd.MarkFlagRequired("port")
	createListenerCmd.MarkFlagRequired("protocol")
	createListenerCmd.MarkFlagRequired("target-group-arn")

	// Update Listener Flags
	updateListenerCmd.Flags().String("load-balancer-id", "", "ID of the load balancer")
	updateListenerCmd.Flags().String("listener-id", "", "ID of the listener")
	updateListenerCmd.Flags().Int("port", 0, "Port number")
	updateListenerCmd.Flags().String("protocol", "", "Protocol (HTTP/HTTPS)")
	updateListenerCmd.Flags().String("target-group-arn", "", "Target group ARN")
	updateListenerCmd.Flags().String("input-json", "", "JSON file containing update configuration")
	updateListenerCmd.MarkFlagRequired("load-balancer-id")
	updateListenerCmd.MarkFlagRequired("listener-id")

	// Get Listener Flags
	getListenerCmd.Flags().String("load-balancer-id", "", "ID of the load balancer")
	getListenerCmd.Flags().String("listener-id", "", "ID of the listener")
	getListenerCmd.MarkFlagRequired("load-balancer-id")
	getListenerCmd.MarkFlagRequired("listener-id")

	// Delete Listener Flags
	deleteListenerCmd.Flags().String("load-balancer-id", "", "ID of the load balancer")
	deleteListenerCmd.Flags().String("listener-id", "", "ID of the listener")
	deleteListenerCmd.MarkFlagRequired("load-balancer-id")
	deleteListenerCmd.MarkFlagRequired("listener-id")
}

var listListenersCmd = &cobra.Command{
	Use:   "list-listeners",
	Short: "List all listeners for a load balancer",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := createClient(cmd.Context())
		if err != nil {
			return err
		}

		lbID, _ := cmd.Flags().GetString("load-balancer-id")
		listeners, _, err := client.ListListeners(cmd.Context(), lbID, 100, "")
		if err != nil {
			return err
		}

		if output == "json" {
			return json.NewEncoder(os.Stdout).Encode(listeners)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tPORT\tPROTOCOL\tTARGET GROUP")
		for _, l := range listeners {
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", l.ID, l.Port, l.Protocol, l.TargetGroupARN)
		}
		return w.Flush()
	},
}

var getListenerCmd = &cobra.Command{
	Use:   "get-listener",
	Short: "Get details of a specific listener",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := createClient(cmd.Context())
		if err != nil {
			return err
		}

		lbID, _ := cmd.Flags().GetString("load-balancer-id")
		listenerID, _ := cmd.Flags().GetString("listener-id")

		listener, err := client.GetListener(cmd.Context(), lbID, listenerID)
		if err != nil {
			return err
		}

		if output == "json" {
			return json.NewEncoder(os.Stdout).Encode(listener)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tPORT\tPROTOCOL\tTARGET GROUP")
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", listener.ID, listener.Port, listener.Protocol, listener.TargetGroupARN)
		return w.Flush()
	},
}

var createListenerCmd = &cobra.Command{
	Use:   "create-listener",
	Short: "Create a new listener",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := createClient(cmd.Context())
		if err != nil {
			return err
		}

		lbID, _ := cmd.Flags().GetString("load-balancer-id")
		var input hlb.ListenerCreate

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
			port, _ := cmd.Flags().GetInt("port")
			protocol, _ := cmd.Flags().GetString("protocol")
			targetGroupARN, _ := cmd.Flags().GetString("target-group-arn")
			certSecretsName, _ := cmd.Flags().GetString("certificate-secrets-name")
			alpnPolicy, _ := cmd.Flags().GetString("alpn-policy")
			enableDeletionProtection, _ := cmd.Flags().GetBool("enable-deletion-protection")

			input = hlb.ListenerCreate{
				Port:                     port,
				Protocol:                 protocol,
				TargetGroupARN:           targetGroupARN,
				CertificateSecretsName:   certSecretsName,
				ALPNPolicy:               alpnPolicy,
				EnableDeletionProtection: enableDeletionProtection,
			}
		}

		listener, err := client.CreateListener(cmd.Context(), lbID, &input)
		if err != nil {
			return err
		}

		if output == "json" {
			return json.NewEncoder(os.Stdout).Encode(listener)
		}

		fmt.Printf("Created listener: %s (Port: %d)\n", listener.ID, listener.Port)
		return nil
	},
}

var updateListenerCmd = &cobra.Command{
	Use:   "update-listener",
	Short: "Update an existing listener",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := createClient(cmd.Context())
		if err != nil {
			return err
		}

		lbID, _ := cmd.Flags().GetString("load-balancer-id")
		listenerID, _ := cmd.Flags().GetString("listener-id")
		var input hlb.ListenerUpdate

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
			if port, _ := cmd.Flags().GetInt("port"); port != 0 {
				input.Port = &port
			}
			if protocol, _ := cmd.Flags().GetString("protocol"); protocol != "" {
				input.Protocol = &protocol
			}
			if targetGroupARN, _ := cmd.Flags().GetString("target-group-arn"); targetGroupARN != "" {
				input.TargetGroupARN = &targetGroupARN
			}
		}

		listener, err := client.UpdateListener(cmd.Context(), lbID, listenerID, &input)
		if err != nil {
			return err
		}

		if output == "json" {
			return json.NewEncoder(os.Stdout).Encode(listener)
		}

		fmt.Printf("Updated listener: %s (Port: %d)\n", listener.ID, listener.Port)
		return nil
	},
}

var deleteListenerCmd = &cobra.Command{
	Use:   "delete-listener",
	Short: "Delete a listener",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := createClient(cmd.Context())
		if err != nil {
			return err
		}

		lbID, _ := cmd.Flags().GetString("load-balancer-id")
		listenerID, _ := cmd.Flags().GetString("listener-id")

		if err := client.DeleteListener(cmd.Context(), lbID, listenerID); err != nil {
			return err
		}

		if output == "json" {
			fmt.Println("{\"status\": \"deleted\"}")
		} else {
			fmt.Printf("Deleted listener: %s\n", listenerID)
		}

		return nil
	},
}
