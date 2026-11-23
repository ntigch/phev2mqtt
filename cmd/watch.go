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
	"encoding/hex"

	"github.com/ntigch/phev2mqtt/protocol"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Connect to Phev and watch incoming updates",
	Long: `Connects to the Phev and watches for incoming register
updates, displaying them in real-time.`,
	SilenceUsage: true,
	RunE:         runWatch,
}

func runWatch(cmd *cobra.Command, args []string) error {
	cl, err := connectClient(cmd)
	if err != nil {
		return err
	}
	defer cl.Close()

	var registers = map[byte]string{}

	for {
		select {
		case m, ok := <-cl.Recv:
			if !ok {
				log.Infof("Connection closed.")
				return nil
			}
			switch m.Type {
			case protocol.CmdInResp:
				dataStr := hex.EncodeToString(m.Data)
				if data := registers[m.Register]; data != dataStr {
					log.Infof("%%PHEV_REG_UPDATE%% %02x: %s -> %s", m.Register, data, dataStr)
					registers[m.Register] = dataStr
					if _, ok := m.Reg.(*protocol.RegisterGeneric); !ok {
						log.Infof("%%PHEV_REG_UPDATE%% %02x: [%s]", m.Register, m.Reg.String())
					}
				}
				if err := cl.SendMessage(&protocol.PhevMessage{
					Type:     protocol.CmdOutSend,
					Register: m.Register,
					Ack:      protocol.Ack,
					Xor:      m.Xor,
					Data:     []byte{0x0},
				}); err != nil {
					log.Infof("Connection closed.")
					return nil
				}
			}
		}
	}
}

func init() {
	clientCmd.AddCommand(watchCmd)
}
