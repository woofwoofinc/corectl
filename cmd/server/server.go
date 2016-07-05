// Copyright (c) 2016 by António Meireles  <antonio.meireles@reformi.st>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"fmt"

	"github.com/TheNewNormal/corectl/components/common"
	"github.com/TheNewNormal/corectl/components/host/darwin/misc/uuid2ip"
	"github.com/TheNewNormal/corectl/components/host/session"
	"github.com/TheNewNormal/corectl/components/server"
	"github.com/TheNewNormal/corectl/release"
	"github.com/helm/helm/log"
	"github.com/satori/go.uuid"

	"github.com/everdev/mack"
	"github.com/spf13/cobra"
)

var (
	serverStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Starts corectld",
		RunE:  serverStartCommand,
	}
	shutdownCmd = &cobra.Command{
		Use:     "stop",
		Aliases: []string{"shutdown"},
		Short:   "Stops corectld",
		RunE:    shutdownCommand,
	}
	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Shows corectld status",
		RunE:  common.PScommand,
	}
	uuidToMacCmd = &cobra.Command{
		Use: "uuid2mac",
		Short: "returns the MAC address that will assembled from the " +
			"given UUID",
		RunE:   uuidToMacCommand,
		Hidden: true,
	}
)

func uuidToMacCommand(cmd *cobra.Command, args []string) (err error) {
	var macAddr string
	if _, err = uuid.FromString(args[0]); err != nil {
		log.Warn("%s not a valid UUID as it doesn't follow RFC "+
			"4122", args[0])
		// given that we only call this with dats generated with
		// uuid.NewV4().String() ...
		err = fmt.Errorf("Something went very wrong, as we're unable to "+
			"generate a MAC address from the provided UUID (%s). Please fill "+
			"a bug at https://github.com/TheNewNormal/corectl/issues with "+
			"this error and wait there for our feedback...", args[0])
	} else if macAddr, err = uuid2ip.GuestMACfromUUID(args[0]); err == nil {
		fmt.Println(macAddr)
	}
	return
}

func shutdownCommand(cmd *cobra.Command, args []string) (err error) {
	if _, err = server.Daemon.Running(); err != nil {
		return
	}
	_, err = server.RPCQuery("Stop", &server.RPCquery{})
	return
}

func serverStartCommand(cmd *cobra.Command, args []string) (err error) {
	var srv *release.Info

	if srv, err = server.Daemon.Running(); err == nil {
		return fmt.Errorf("corectld already started (with pid %v)",
			srv.Pid)
	}

	if !session.Caller.Privileged {
		if err = mack.Tell("System Events",
			"do shell script \""+
				session.Executable()+" start --user "+session.Caller.Username+
				" > /dev/null 2>&1 & \" with administrator privileges",
			"delay 3"); err != nil {
			return
		}
		if srv, err = server.Daemon.Running(); err != nil {
			return err
		}
		fmt.Println("Started corectld:")
		srv.PrettyPrint(true)
		return
	}
	server.Daemon = server.New()
	server.Daemon.Active = make(map[string]*server.VMInfo)
	return server.Start()
}

func init() {
	serverStartCmd.Flags().StringP("user", "u", "",
		"sets the user that will 'own' the corectld instance")
	serverStartCmd.Flags().BoolP("force", "f", false,
		"rebuilds config drive iso even if a suitable one is already present")
	rootCmd.AddCommand(shutdownCmd, statusCmd, serverStartCmd, uuidToMacCmd)
}
