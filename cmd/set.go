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
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set register value(s)",
	Args:  cobra.MinimumNArgs(1),
	Long: `Send new values to the car to set register(s).

Each arg is of the form <register>:<value> - the given register is set to the
given value. Each should be a hex string.

e.g:

0b:02 will set register 0b to the value 02

`,
	SilenceUsage: true,
	RunE:         runSet,
}

type regValue struct {
	register byte
	value    []byte
}

func runSet(cmd *cobra.Command, args []string) error {
	setRegisters := []*regValue{}
	for _, arg := range args {
		vals := strings.Split(arg, ":")
		if len(vals) != 2 {
			return fmt.Errorf("bad argument %q, want <register>:<value>", arg)
		}
		register, err := hex.DecodeString(vals[0])
		if err != nil || len(register) != 1 {
			return fmt.Errorf("bad register in %q, want one hex byte", arg)
		}
		value, err := hex.DecodeString(vals[1])
		if err != nil {
			return fmt.Errorf("bad value in %q: %v", arg, err)
		}
		setRegisters = append(setRegisters, &regValue{
			register: register[0], value: value,
		})
	}

	waitTime, err := cmd.Flags().GetDuration("wait_duration")
	if err != nil {
		return err
	}
	sendInterval, err := cmd.Flags().GetDuration("send_interval")
	if err != nil {
		return err
	}

	cl, err := connectClient(cmd)
	if err != nil {
		return err
	}
	defer cl.Close()
	log.Infof("Waiting %s", waitTime.String())

	time.Sleep(waitTime)

	for _, reg := range setRegisters {
		log.Infof("Setting register 0x%x to 0x%s", reg.register, hex.EncodeToString(reg.value))
		if err := cl.SetRegister(reg.register, reg.value); err != nil {
			return err
		}
		time.Sleep(sendInterval)
	}
	return nil
}

func init() {
	clientCmd.AddCommand(setCmd)

	setCmd.Flags().Duration("wait_duration", 10*time.Second, "How long to wait after connecting to car before sending register update")
	setCmd.Flags().Duration("send_interval", 1*time.Second, "Interval between setting each register")
}
