// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package namespace

import (
	"github.com/streamnative/pulsarctl/pkg/cmdutils"
	"github.com/streamnative/pulsarctl/pkg/ctl/utils"

	util "github.com/streamnative/pulsarctl/pkg/pulsar/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func setRetention(vc *cmdutils.VerbCmd) {
	desc := cmdutils.LongDescription{}
	desc.CommandUsedFor = "Set the retention policy for a namespace"
	desc.CommandPermission = "This command requires tenant admin permissions."

	var examples []cmdutils.Example
	setRetentionWithTime := cmdutils.Example{
		Desc:    "Set the retention policy for a namespace",
		Command: "pulsarctl namespaces set-retention tenant/namespace --time 100m --size 1G",
	}

	setInfiniteTime := cmdutils.Example{
		Desc:    "Set the infinite time retention policy for a namespace",
		Command: "pulsarctl namespaces set-retention tenant/namespace --size -1",
	}
	examples = append(examples, setRetentionWithTime, setInfiniteTime)
	desc.CommandExamples = examples

	var out []cmdutils.Output
	successOut := cmdutils.Output{
		Desc: "normal output",
		Out:  "Set retention successfully for [tenant/namespace]",
	}

	noNamespaceName := cmdutils.Output{
		Desc: "you must specify a tenant/namespace name, please check if the tenant/namespace name is provided",
		Out:  "[✖]  the namespace name is not specified or the namespace name is specified more than one",
	}

	tenantNotExistError := cmdutils.Output{
		Desc: "the tenant does not exist",
		Out:  "[✖]  code: 404 reason: Tenant does not exist",
	}

	nsNotExistError := cmdutils.Output{
		Desc: "the namespace does not exist",
		Out:  "[✖]  code: 404 reason: Namespace (tenant/namespace) does not exist",
	}

	notSetBacklog := cmdutils.Output{
		Desc: "Retention Quota must exceed configured backlog quota for namespace",
		Out:  "Retention Quota must exceed configured backlog quota for namespace",
	}

	out = append(out, successOut, noNamespaceName, tenantNotExistError, nsNotExistError, notSetBacklog)
	desc.CommandOutput = out

	vc.SetDescription(
		"set-retention",
		"Set the retention policy for a namespace",
		desc.ToString(),
		desc.ExampleToString(),
		"set-retention",
	)

	var data util.NamespacesData

	vc.FlagSetGroup.InFlagSet("Set Retention", func(flagSet *pflag.FlagSet) {
		flagSet.StringVar(
			&data.RetentionTimeStr,
			"time",
			"",
			"Retention time in minutes (or minutes, hours,days,weeks eg: 100m, 3h, 2d, 5w).\n"+
				"0 means no retention and -1 means infinite time retention")

		flagSet.StringVar(
			&data.LimitStr,
			"size",
			"",
			"Retention size limit (eg: 10M, 16G, 3T).\n"+
				"0 or less than 1MB means no retention and -1 means infinite size retention")

		_ = cobra.MarkFlagRequired(flagSet, "time")
		_ = cobra.MarkFlagRequired(flagSet, "size")
	})
	vc.EnableOutputFlagSet()

	vc.SetRunFuncWithNameArg(func() error {
		return doSetRetention(vc, data)
	}, "the namespace name is not specified or the namespace name is specified more than one")
}

func doSetRetention(vc *cmdutils.VerbCmd, data util.NamespacesData) error {
	ns := vc.NameArg
	admin := cmdutils.NewPulsarClient()
	sizeLimit, err := utils.ValidateSizeString(data.LimitStr)
	if err != nil {
		return err
	}
	retentionTimeInSecond, err := utils.ParseRelativeTimeInSeconds(data.RetentionTimeStr)
	if err != nil {
		return err
	}

	var (
		retentionTimeInMin int
		retentionSizeInMB  int
	)

	if retentionTimeInSecond != -1 {
		retentionTimeInMin = int(retentionTimeInSecond.Minutes())
	} else {
		retentionTimeInMin = -1
	}

	if sizeLimit != -1 {
		retentionSizeInMB = int(sizeLimit / (1024 * 1024))
	} else {
		retentionSizeInMB = -1
	}
	err = admin.Namespaces().SetRetention(ns, util.NewRetentionPolicies(retentionTimeInMin, retentionSizeInMB))
	if err == nil {
		vc.Command.Printf("Set retention successfully for [%s]."+
			" The retention policy is: time = %d min, size = %d MB\n", ns, retentionTimeInMin, retentionSizeInMB)
	}

	return err
}
