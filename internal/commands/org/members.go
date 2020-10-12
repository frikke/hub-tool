/*
   Copyright 2020 Docker Hub Tool authors

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

package org

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"

	"github.com/docker/hub-cli-plugin/internal/format"
	"github.com/docker/hub-cli-plugin/internal/hub"
	"github.com/docker/hub-cli-plugin/internal/metrics"
)

const (
	membersName = "members"
)

var (
	memberColumns = []memberColumn{
		{"USERNAME", func(m hub.Member) string { return m.Username }},
		{"FULL NAME", func(m hub.Member) string { return m.FullName }},
	}
)

type memberColumn struct {
	header string
	value  func(m hub.Member) string
}

type memberOptions struct {
	format.Option
}

func newMembersCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts memberOptions
	cmd := &cobra.Command{
		Use:   membersName + " ORGANIZATION",
		Short: "List all the members in an organization",
		Args:  cli.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, membersName)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMembers(streams, hubClient, opts, args[0])
		},
	}
	opts.AddFormatFlag(cmd.Flags())
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func runMembers(streams command.Streams, hubClient *hub.Client, opts memberOptions, organization string) error {
	members, err := hubClient.GetMembers(organization)
	if err != nil {
		return err
	}
	return opts.Print(streams.Out(), members, printMembers)
}

func printMembers(out io.Writer, values interface{}) error {
	members := values.([]hub.Member)
	w := tabwriter.NewWriter(out, 20, 1, 3, ' ', 0)
	var headers []string
	for _, column := range memberColumns {
		headers = append(headers, column.header)
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	for _, member := range members {
		var values []string
		for _, column := range memberColumns {
			values = append(values, column.value(member))
		}
		fmt.Fprintln(w, strings.Join(values, "\t"))
	}
	return w.Flush()
}
