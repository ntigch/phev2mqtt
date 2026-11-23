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
	"fmt"
	"os"
	"strings"

	"github.com/ntigch/phev2mqtt/protocol"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var securityKey *protocol.SecurityKey

// fileCmd represents the file command
var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Decode Phev messages from a file",
	Long: `
Decode raw hex string messages from the provided filename.
`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		inData, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}
		// Ignore line breaks and other whitespace between hex digits.
		dat := strings.Join(strings.Fields(string(inData)), "")
		binData, err := hex.DecodeString(dat)
		if err != nil {
			return fmt.Errorf("%s: %w", args[0], err)
		}
		securityKey = &protocol.SecurityKey{}
		for _, msg := range protocol.NewFromBytes(binData, securityKey) {
			log.Debug(hex.EncodeToString(msg.Original))
			log.Infof("%s", msg.ShortForm())
		}
		return nil
	},
}

func init() {
	decodeCmd.AddCommand(fileCmd)
}
