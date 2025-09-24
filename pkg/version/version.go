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

package version

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	GitSHA    = "build ./build.sh"
	GitTag    = ""
	BuildDate = ""
)

func ReleaseVersion() string {
	var version string

	if GitTag != "" {
		version = GitTag
	}

	if GitSHA != "" {
		version += fmt.Sprintf("-%s", GitSHA)
	}

	if BuildDate != "" {
		version += fmt.Sprintf("-%s", BuildDate)
	}

	if version == "" {
		version = "latest"
	}

	return version
}

func GoV() string {
	v := strings.TrimPrefix(runtime.Version(), "go")
	if strings.Count(v, ".") > 1 {
		v = v[:strings.LastIndex(v, ".")]
	}
	return v
}

func GetVersionTemplate() string {
	var tpl string
	tpl += fmt.Sprintf("maco Version: %s\n", GitTag)
	tpl += fmt.Sprintf("Git SHA: %s\n", GitSHA)
	tpl += fmt.Sprintf("Go Version: %s\n", runtime.Version())
	tpl += fmt.Sprintf("Go OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	return tpl
}
