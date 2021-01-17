package net

import (
	"encoding/json"
	"fmt"
	"fnd/cli"
	"fnd/rpc"
	apiv1 "fnd/rpc/v1"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type peerJSON struct {
	ID          string `json:"id"`
	IP          string `json:"ip"`
	Banned      bool   `json:"banned"`
	Whitelisted bool   `json:"whitelisted"`
	Connected   bool   `json:"connected"`
	TxBytes     int    `json:"tx_bytes"`
	RxBytes     int    `json:"rx_bytes"`
}

var peerInfoCmd = &cobra.Command{
	Use:   "peer-info",
	Short: "Returns information about all peers DDRP has heard of.",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := cli.DialRPC(cmd)
		if err != nil {
			return err
		}
		grpcClient := apiv1.NewDDRPv1Client(conn)
		peers, err := rpc.ListPeers(grpcClient)
		if err != nil {
			return err
		}

		format, _ := cmd.Flags().GetString(cli.FlagFormat)
		if format == "json" {
			encoder := json.NewEncoder(os.Stdout)
			for _, peer := range peers {
				jsonPeer := &peerJSON{
					ID:          peer.ID,
					IP:          peer.IP,
					Banned:      peer.Banned,
					Whitelisted: peer.Whitelisted,
					Connected:   peer.Connected,
					TxBytes:     int(peer.TxBytes),
					RxBytes:     int(peer.RxBytes),
				}

				if err := encoder.Encode(jsonPeer); err != nil {
					return err
				}
			}
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{
				"Peer ID",
				"IP",
				"Banned",
				"Whitelisted",
				"Connected",
				"Tx Bytes",
				"Rx Bytes",
			})
			for _, res := range peers {
				table.Append([]string{
					res.ID,
					res.IP,
					boolToStr(res.Banned),
					boolToStr(res.Whitelisted),
					boolToStr(res.Connected),
					bandwidthToStr(res.TxBytes),
					bandwidthToStr(res.RxBytes),
				})
			}

			table.Render()
			fmt.Println("")
			fmt.Printf("Total: %d\n", len(peers))
		}
		return nil
	},
}

func bandwidthToStr(stat uint64) string {
	if stat == 0 {
		return "-"
	}

	unit := uint64(1000)
	if stat < unit {
		return fmt.Sprintf("%d B", stat)
	}
	div, exp := unit, 0
	for n := stat / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(stat)/float64(div), "kMGTPE"[exp])
}

func boolToStr(val bool) string {
	if val {
		return "TRUE"
	}
	return "FALSE"
}

func init() {
	cmd.AddCommand(peerInfoCmd)
}
