package pocketbase

import (
	"bytes"
	"os"
	"testing"

	"github.com/pluja/pocketbase/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackup_FullList(t *testing.T) {
	t.Run("without authorization", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)
		resp, err := defaultClient.Backup().FullList()
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "valid record authorization")
	})

	t.Run("with valid authentication, but no existing backups", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL, WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		resp, err := defaultClient.Backup().FullList()
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Empty(t, resp)
	})

	t.Run("with valid authentication, create backup and check", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL, WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		err := defaultClient.Backup().Create()
		require.NoError(t, err)

		resp, err := defaultClient.Backup().FullList()
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp)

		// cleanup
		_ = defaultClient.Backup().Delete(resp[0].Key)
	})
}

func TestBackup_Create(t *testing.T) {
	t.Run("without authorization", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)
		err := defaultClient.Backup().Create()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "valid record authorization")
	})

	t.Run("with valid authentication, create backup without name and check", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL, WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		resp, err := defaultClient.Backup().FullList()
		require.NoError(t, err)
		require.Empty(t, resp)

		err = defaultClient.Backup().Create()
		assert.NoError(t, err)

		resp, err = defaultClient.Backup().FullList()
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp)

		// cleanup
		_ = defaultClient.Backup().Delete(resp[0].Key)
	})

	t.Run("with valid authentication, create backup with name and check", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL, WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		const backupName = "foobar"
		err := defaultClient.Backup().Create(backupName)
		assert.NoError(t, err)

		assert.True(t, isBackupexisting(t, defaultClient, backupName+".zip"))

		// cleanup
		_ = defaultClient.Backup().Delete(backupName + ".zip")
	})

	t.Run("with valid authentication, create backup with name incl. zip-extension and check", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL, WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		const backupName = "barfoo.zip"
		err := defaultClient.Backup().Create(backupName)
		assert.NoError(t, err)

		assert.True(t, isBackupexisting(t, defaultClient, backupName))

		// cleanup
		_ = defaultClient.Backup().Delete(backupName)
	})
}

func TestBackup_Delete(t *testing.T) {
	t.Run("without authorization", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)
		err := defaultClient.Backup().Delete("foobar")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "valid record authorization")
	})

	t.Run("create a backup and delete it", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL, WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		backupName := "foobar.zip"

		err := defaultClient.Backup().Create(backupName)
		require.NoError(t, err)
		require.True(t, isBackupexisting(t, defaultClient, backupName))

		err = defaultClient.Backup().Delete(backupName)
		assert.NoError(t, err)
		assert.False(t, isBackupexisting(t, defaultClient, backupName))
	})

	t.Run("create a backup and delete a non existing", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL, WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		backupName := "foobar.zip"

		err := defaultClient.Backup().Create(backupName)
		require.NoError(t, err)
		require.True(t, isBackupexisting(t, defaultClient, backupName))

		err = defaultClient.Backup().Delete(backupName + backupName)
		assert.Error(t, err)
		assert.False(t, isBackupexisting(t, defaultClient, backupName+backupName))

		// cleanup
		require.NoError(t, defaultClient.Backup().Delete(backupName))
	})
}

func TestBackup_Restore(t *testing.T) {
	t.Run("without authorization", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)
		err := defaultClient.Backup().Restore("foobar")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "valid record authorization")
	})

	t.Run("restore a backup", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL, WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		backupName := "foobar.zip"

		err := defaultClient.Backup().Create(backupName)
		require.NoError(t, err)
		require.True(t, isBackupexisting(t, defaultClient, backupName))

		err = defaultClient.Backup().Restore(backupName)
		assert.NoError(t, err)

		// cleanup
		require.NoError(t, defaultClient.Backup().Delete(backupName))
	})

	t.Run("cannot restore a backup with a nox existing key", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL, WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		backupName := "not_existing.zip"

		err := defaultClient.Backup().Restore(backupName)
		assert.Error(t, err)
	})
}

func TestBackup_Upload(t *testing.T) {
	t.Run("without authorization", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)
		err := defaultClient.Backup().Upload("foobar", bytes.NewReader([]byte{10}))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "valid record authorization")
	})

	t.Run("upload a backup", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL, WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		backupName := "foobar_test.zip"

		file, err := os.Open("./testressources/pb_backup.zip")
		require.NoError(t, err)
		defer file.Close()

		err = defaultClient.Backup().Upload(backupName, file)
		require.NoError(t, err)
		require.True(t, isBackupexisting(t, defaultClient, backupName))

		err = defaultClient.Backup().Restore(backupName)
		assert.NoError(t, err)

		// cleanup
		require.NoError(t, defaultClient.Backup().Delete(backupName))
	})
}

func TestBackup_GetDownloadURL(t *testing.T) {
	t.Run("build URL with token and key", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL, WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		url, err := defaultClient.Backup().GetDownloadURL("token", "key")
		assert.NoError(t, err)
		assert.Equal(t, defaultURL+"/api/backups/key?token=token", url)
	})

	t.Run("build URL with no token and no key", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL, WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		url, err := defaultClient.Backup().GetDownloadURL("", "")
		assert.Error(t, err)
		assert.Empty(t, url)
	})
}

func isBackupexisting(t *testing.T, defaultClient *Client, backupname string) bool {
	resp, err := defaultClient.Backup().FullList()
	require.NoError(t, err)

	for _, b := range resp {
		if b.Key == backupname {
			return true
		}
	}
	return false
}
