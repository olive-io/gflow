/*
Copyright 2025 The gflow Authors

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

package app

import (
	"fmt"
	"time"
)

type OutputFormat string

const (
	OutputJSON   OutputFormat = "json"
	OutputTable  OutputFormat = "table"
	OutputSimple OutputFormat = "simple"
)

type Printer struct {
	format OutputFormat
}

func NewPrinter(format OutputFormat) *Printer {
	return &Printer{format: format}
}

func (p *Printer) Print(data any) error {
	switch p.format {
	case OutputJSON:
		return p.printJSON(data)
	default:
		return p.printSimple(data)
	}
}

func (p *Printer) printJSON(data any) error {
	fmt.Println(data)
	return nil
}

func (p *Printer) printSimple(data any) error {
	fmt.Println(data)
	return nil
}

func (p *Printer) PrintTable(headers []string, rows [][]string) error {
	for _, h := range headers {
		fmt.Printf("%-20s ", h)
	}
	fmt.Println()
	for _, row := range rows {
		for _, cell := range row {
			fmt.Printf("%-20s ", cell)
		}
		fmt.Println()
	}
	return nil
}

func FormatTimestamp(ts int64) string {
	if ts == 0 {
		return "-"
	}
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}
