// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package storebenchmark

import (
	"os"
)

func getEnv(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	} else {
		return defaultValue
	}
}
