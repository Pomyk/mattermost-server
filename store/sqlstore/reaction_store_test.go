// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"testing"

	"github.com/mattermost/mattermost-server/v5/store/storebenchmark"
	"github.com/mattermost/mattermost-server/v5/store/storetest"
)

func TestReactionStore(t *testing.T) {
	StoreTest(t, storetest.TestReactionStore)
}

func BenchmarkReactionStore(b *testing.B) {
	StoreBenchmark(b, storebenchmark.BenchmarkReactionStore)
}
