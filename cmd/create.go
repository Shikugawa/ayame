// Copyright 2021 Rei Shimizu

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package cmd

import (
	"io/ioutil"

	"github.com/Shikugawa/ayame/pkg/config"
	"github.com/Shikugawa/ayame/pkg/state"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	configPath string

	createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create network environment from config",
		Run: func(cmd *cobra.Command, args []string) {
			bytes, err := ioutil.ReadFile(configPath)
			if err != nil {
				log.Errorf(err.Error())
				return
			}

			cfg, err := config.ParseConfig(bytes)
			if err != nil {
				log.Errorf(err.Error())
				return
			}

			st, _ := state.LoadStateFromFile()
			s, err := state.InitAll(cfg, st, false)
			if err != nil {
				log.Errorf(err.Error())
				return
			}

			log.Info("succeeded to initialize")

			if err := s.SaveState(); err != nil {
				log.Errorf(err.Error())
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&configPath, "config", "c", "", "config path")
	createCmd.MarkFlagRequired("config")
}
