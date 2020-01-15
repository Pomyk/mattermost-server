// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package migration

import (
	"testing"

	"github.com/mattermost/mattermost-server/v5/store/sqlstore"
	"github.com/mattermost/mattermost-server/v5/store/storetest/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCreateIndexMocked(t *testing.T) {
	mockSupplier := &mocks.SqlStore{}
	runner := NewMigrationRunner(mockSupplier)

	createIndex := NewCreateIndex("posts_root_id_delete_at", "Posts", []string{"RootId", "DeleteAt"}, sqlstore.INDEX_TYPE_DEFAULT, false)
	err := runner.Add(createIndex)
	assert.Nil(t, err, "should have added migration")
}
