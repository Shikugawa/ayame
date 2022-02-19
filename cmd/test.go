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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Shikugawa/ayame/pkg/config"
	"github.com/Shikugawa/ayame/pkg/state"
	"github.com/r3labs/diff"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	dataName = "config.yml"
	refName  = "state.json"
)

var (
	datasetPath string
	test        string

	testCmd = &cobra.Command{
		Use:   "test",
		Short: "run tests",
		Run: func(cmd *cobra.Command, args []string) {
			testdata := make(map[string][]string)

			if err := filepath.Walk(datasetPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					if _, ok := testdata[info.Name()]; !ok {
						testdata[info.Name()] = []string{}
					}
					return nil
				}

				splited := strings.Split(path, "/")

				if len(splited) < 2 {
					return nil
				}

				fileName := splited[len(splited)-1]
				parentDir := splited[len(splited)-2]

				if _, ok := testdata[parentDir]; ok {
					if fileName == dataName || fileName == refName {
						testdata[parentDir] = append(testdata[parentDir], path)
					}
				}

				return nil
			}); err != nil {
				log.Errorf(err.Error())
			}

			for testName, paths := range testdata {
				if len(paths) == 0 {
					continue
				}

				if len(test) != 0 && test != testName {
					continue
				}

				splited := strings.Split(testName, "-")
				if len(splited) != 2 {
					log.Errorf("invalid testname format: %s", testName)
					continue
				}

				shouldSuccess := splited[1] == "ok"

				var cfg []byte
				var ref []byte

				for _, path := range paths {
					if strings.HasSuffix(path, "/"+dataName) {
						bytes, err := ioutil.ReadFile(path)
						if err != nil {
							log.Errorf("failed to read file: %s", path)
							continue
						}
						cfg = bytes
						continue
					}

					if strings.HasSuffix(path, "/"+refName) {
						bytes, err := ioutil.ReadFile(path)
						if err != nil {
							log.Errorf("failed to read file: %s", path)
							continue
						}
						ref = bytes
						continue
					}
				}

				if len(cfg) != 0 && len(ref) != 0 {
					log.Infof("================ start test: %s ================", testName)

					c, err := config.ParseConfig(cfg)
					if err != nil {
						log.Errorf(err.Error())
						continue
					}

					s, err := state.InitResources(c, true)
					if err != nil {
						if !shouldSuccess {
							log.Infof("failed with error: %s", err.Error())
							log.Infof("================ test %s OK ================", testName)
						} else {
							log.Errorf(err.Error())
						}

						continue
					}

					ls, err := s.DumpAll()
					if err != nil {
						log.Errorf(err.Error())
						continue
					}

					expectedState := state.LoadStateFromBytes(ref)
					if expectedState == nil {
						continue
					}

					lsRef, err := expectedState.DumpAll()
					if err != nil {
						log.Errorf(err.Error())
						continue
					}

					change, err := diff.Diff(ls, lsRef)
					if err != nil {
						log.Errorf(err.Error())
						continue
					}

					if len(change) != 0 {
						log.Errorln(change)
					} else {
						log.Infof("================ test %s OK ================", testName)
					}

				} else {
					log.Errorf("================ failed to start test: %s ================", testName)
				}
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(testCmd)

	testCmd.Flags().StringVarP(&datasetPath, "path", "p", "", "test dataset path")
	testCmd.MarkFlagRequired("path")

	testCmd.Flags().StringVarP(&test, "test", "t", "", "target test")
}
