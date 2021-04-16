/*
Copyright © 2021 Rei Shimizu

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/Shikugawa/ayame/pkg/state"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "get current resources",
	Run: func(cmd *cobra.Command, args []string) {
		s, err := state.LoadStateFromFile()
		if err != nil {
			log.Fatalln(err)
			return
		}

		ls, err := s.DumpAll()
		if err != nil {
			log.Fatalln(err)
			return
		}

		fmt.Println(ls)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
