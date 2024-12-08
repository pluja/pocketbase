package pocketbase

import (
	"strings"
	"testing"
	"time"

	"github.com/pluja/pocketbase/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type User struct {
	AuthProviders    []interface{} `json:"authProviders"`
	UsernamePassword bool          `json:"usernamePassword"`
	EmailPassword    bool          `json:"emailPassword"`
	OnlyVerified     bool          `json:"onlyVerified"`
}

func TestCollection_ListAuthMethods(t *testing.T) {
	t.Run("get AuthMethods with invalid authorization", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL, WithAdminEmailPassword("foo", "bar"))

		resp, err := CollectionSet[User](defaultClient, "users").ListAuthMethods()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to authenticate.")
		assert.Empty(t, resp)
	})

	t.Run("get AuthMethods with valid authorization", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL, WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))

		resp, err := CollectionSet[User](defaultClient, "users").ListAuthMethods()
		assert.NoError(t, err)
		assert.True(t, resp.Password.Enabled)
		assert.False(t, resp.OAuth2.Enabled)
		assert.False(t, resp.MFA.Enabled)
		assert.False(t, resp.OTP.Enabled)
		assert.Empty(t, resp.AuthProviders)
		assert.False(t, resp.UsernamePassword)
		assert.True(t, resp.EmailPassword)
	})
}

func TestCollection_AuthWithPassword(t *testing.T) {
	t.Run("authenticate with valid user credentials", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)

		response, err := CollectionSet[User](defaultClient, "users").AuthWithPassword("user@user.com", "user@user.com")
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Token)
		assert.Len(t, response.Token, 224)
		assert.Equal(t, response.Token, defaultClient.token)
	})

	t.Run("authenticate with invalid user credentials", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)

		response, err := CollectionSet[User](defaultClient, "users").AuthWithPassword("foo", "bar")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to authenticate")
		assert.Empty(t, response.Token)
		assert.Len(t, response.Token, 0)
		assert.Equal(t, response.Token, defaultClient.token)
	})
}

func TestCollection_AuthWithOauth2(_ *testing.T) {
	// actually I don't know how to test
}

func TestCollection_AuthRefresh(t *testing.T) {
	t.Run("refresh authentication without valid user auth token", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)

		_, err := CollectionSet[User](defaultClient, "users").AuthRefresh()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "valid record authorization")
	})

	t.Run("refresh authentication with invalid user auth token", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)

		defaultClient.token = strings.Repeat("X", 207)
		_, err := CollectionSet[User](defaultClient, "users").AuthRefresh()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "valid record authorization")
	})

	t.Run("refresh authentication with valid user auth token", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)

		authResponse, err := CollectionSet[User](defaultClient, "users").AuthWithPassword("user@user.com", "user@user.com")
		require.NoError(t, err)
		require.NotEmpty(t, authResponse.Token)
		oldToken := authResponse.Token

		time.Sleep(1 * time.Second) // we need to wait to get another token expire time

		response, err := CollectionSet[User](defaultClient, "users").AuthRefresh()
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Token)
		assert.Len(t, response.Token, 207)
		assert.Equal(t, response.Token, defaultClient.token)
		assert.NotEqual(t, response.Token, oldToken)
	})
}

func TestCollection_RequestVerification(t *testing.T) {
	t.Run("request verification with valid authorization and not existing user", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL, WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))

		err := CollectionSet[User](defaultClient, "users").RequestVerification("nouser@nouser.com")
		assert.NoError(t, err)
	})

	t.Run("request verification with valid authorization", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL, WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))

		err := CollectionSet[User](defaultClient, "users").RequestVerification("user@user.com")
		assert.NoError(t, err)
	})
}

func TestCollection_ConfirmVerification(t *testing.T) {
	t.Run("confirm verification with an invalid verification token", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)

		err := CollectionSet[User](defaultClient, "users").ConfirmVerification("no-valid-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation_invalid_token_claims")
	})

	t.Run("confirm verification with an valid token but not for the test-environment verification token", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)

		err := CollectionSet[User](defaultClient, "users").ConfirmVerification("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2xsZWN0aW9uSWQiOiJfcGJfdXNlcnNfYXV0aF8iLCJlbWFpbCI6InVzZXJAdXNlci5jb20iLCJleHAiOjE3MTQwNzE0MzgsImlkIjoiOHZ4OWh1ZDZkZXAyMnV2IiwidHlwZSI6ImF1dGhSZWNvcmQifQ.UwHOhmd0F_kK4LdjvDYqzE7QMheXmIiipFM6i-gwEPQ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation_invalid_token")
	})
}

func TestCollection_RequestPasswordReset(t *testing.T) {
	t.Run("request password reset with valid authorization and not existing user", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)

		err := CollectionSet[User](defaultClient, "users").RequestPasswordReset("nouser@nouser.com")
		assert.NoError(t, err)
	})

	t.Run("request password reset with valid authorization", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)

		err := CollectionSet[User](defaultClient, "users").RequestPasswordReset("user@user.com")
		assert.NoError(t, err)
	})
}

func TestCollection_ConfirmPassworReset(t *testing.T) {
	t.Run("confirm password reset with an invalid verification token", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)

		err := CollectionSet[User](defaultClient, "users").ConfirmPasswordReset("no-valid-token", "new-password-123", "new-password-123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation_invalid_token")
	})

	t.Run("confirm password reset with an valid token but not for the test-environment verification token", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)

		err := CollectionSet[User](defaultClient, "users").ConfirmPasswordReset("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2xsZWN0aW9uSWQiOiJfcGJfdXNlcnNfYXV0aCIsImVtYWlsIjoidXNlckB1c2VyLmNvbSIsImV4cCI6MTcxMzQ3MTc5NSwiaWQiOiI4dng5aHVkNmRlcDIydXYiLCJ0eXBlIjoiYXV0aFJlY29yZCJ9.u_7_u1t0MueFfKAMmXPqe4o1mNBn_-oFEpdSSeGqlUs",
			"new-password-123",
			"new-password-123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation_invalid_token")
	})
}

func TestCollection_RequestEmailChange(t *testing.T) {
	t.Run("confirm pemail change without a valid login", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)

		err := CollectionSet[User](defaultClient, "users").RequestEmailChange("useruser@user.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "valid record authorization")
	})
}

func TestCollection_ConfirmEmailChange(t *testing.T) {
	t.Run("confirm email change with an invalid verification token", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)

		err := CollectionSet[User](defaultClient, "users").ConfirmEmailChange("no-valid-token", "new-password-123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation_invalid_token_payload")
	})

	t.Run("confirm email change with an valid token but not for the test-environment verification token", func(t *testing.T) {
		defaultClient := NewTestClient(defaultURL)

		err := CollectionSet[User](defaultClient, "users").ConfirmEmailChange("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2xsZWN0aW9uSWQiOiJfcGJfdXNlcnNfYXV0aCIsImVtYWlsIjoidXNlckB1c2VyLmNvbSIsImV4cCI6MTcxMzQ3MTc5NSwiaWQiOiI4dng5aHVkNmRlcDIydXYiLCJ0eXBlIjoiYXV0aFJlY29yZCJ9.u_7_u1t0MueFfKAMmXPqe4o1mNBn_-oFEpdSSeGqlUs",
			"user@user.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation_invalid_token")
	})
}
