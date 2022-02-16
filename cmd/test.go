// Copyright 2022 Rei Shimizu

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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Shikugawa/ayame/pkg/config"
	"github.com/Shikugawa/ayame/pkg/state"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	datasetPath string
	testCmd     = &cobra.Command{
		Use:   "test",
		Short: "run tests",
		Run: func(cmd *cobra.Command, args []string) {
			if err := filepath.Walk(datasetPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				log.Infof("Run Test: %s", path)
				bytes, err := ioutil.ReadFile(path)

				if err != nil {
					log.Errorf(err.Error())
					return nil
				}

				cfg, err := config.ParseConfig(bytes)
				if err != nil {
					log.Errorf(err.Error())
					return nil
				}

				st, _ := state.LoadStateFromFile()
				s, err := state.InitAll(cfg, st, true)
				if err != nil {
					log.Errorf(err.Error())
					return nil
				}

				ls, err := s.DumpAll()
				if err != nil {
					log.Errorf(err.Error())
					return nil
				}

				fmt.Println(ls)
				return nil
			}); err != nil {
				log.Errorf(err.Error())
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(testCmd)

	testCmd.Flags().StringVarP(&datasetPath, "path", "p", "", "test dataset path")
	testCmd.MarkFlagRequired("path")
}
