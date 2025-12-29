// Copyright (c) JFrog Ltd. (2025)
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

package artifactory

import (
	"fmt"

	"github.com/samber/lo"
)

type artifactoryError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e artifactoryError) String() string {
	return fmt.Sprintf("%s - %s", e.Code, e.Message)
}

type ArtifactoryErrorsResponse struct {
	Errors []artifactoryError `json:"errors"`
}

func (r ArtifactoryErrorsResponse) String() string {
	errs := lo.Reduce(r.Errors, func(err string, item artifactoryError, _ int) string {
		if err == "" {
			return item.String()
		} else {
			return fmt.Sprintf("%s, %s", err, item.String())
		}
	}, "")
	return errs
}
