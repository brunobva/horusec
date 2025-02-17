// Copyright 2020 ZUP IT SERVICOS EM TECNOLOGIA E INOVACAO SA
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

package entities

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ZupIT/horusec/development-kit/pkg/enums/severity"
)

type Result struct {
	File       string `json:"file"`
	Line       string `json:"line"`
	Column     string `json:"column"`
	Level      string `json:"level"`
	Warning    string `json:"warning"`
	Suggestion string `json:"suggestion"`
	Note       string `json:"note"`
	Context    string `json:"context"`
}

func (r *Result) GetDetails() string {
	return fmt.Sprintf("%s %s %s", r.Warning, r.Suggestion, r.Note)
}

func (r *Result) GetSeverity() severity.Severity {
	level, _ := strconv.Atoi(r.Level)
	return r.mapSeverityByLevels()[level]
}

func (r *Result) mapSeverityByLevels() map[int]severity.Severity {
	return map[int]severity.Severity{
		5: severity.Critical,
		4: severity.High,
		3: severity.Medium,
		2: severity.Medium,
		1: severity.Low,
		0: severity.Low,
	}
}

func (r *Result) GetFilename() string {
	return strings.ReplaceAll(r.File, "./", "")
}
