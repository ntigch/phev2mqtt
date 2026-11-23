/*
Copyright © 2021 Ben Buxton <bbuxton@gmail.com>
Copyright © 2026 ntigch

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"github.com/ntigch/phev2mqtt/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Client to connect and interact with the vehicle",
	Long: `
Client functionality for connecting and interacting with
the vehicle. Please see subcommands for available functions.
Requires Wifi reachability to the vehicle.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)

	clientCmd.PersistentFlags().String("address", "192.168.8.46:8080", "Address to connect to")

	viper.BindPFlag("address", clientCmd.PersistentFlags().Lookup("address"))
}

// connectClient connects to the car at the --address flag and waits for it
// to start the session.
func connectClient(cmd *cobra.Command) (*client.Client, error) {
	address, _ := cmd.Flags().GetString("address")
	cl, err := client.New(client.AddressOption(address))
	if err != nil {
		return nil, err
	}
	if err := cl.Connect(); err != nil {
		return nil, err
	}
	if err := cl.Start(); err != nil {
		cl.Close()
		return nil, err
	}
	log.Infof("Client connected and started!")
	return cl, nil
}
