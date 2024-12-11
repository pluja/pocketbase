package pocketbase

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pluja/pocketbase/migrations"
)

const (
	defaultURL = "http://127.0.0.1:8090"
)

// REMEMBER to start the Pocketbase before running this example with `make serve` command

func TestAuthorizeAnonymous(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Empty credentials",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(defaultURL)
			err := c.Authorize()
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestAuthorizeEmailPassword(t *testing.T) {
	type args struct {
		email    string
		password string
	}
	tests := []struct {
		name    string
		admin   args
		user    args
		wantErr bool
	}{
		{
			name:    "Valid credentials admin",
			admin:   args{email: migrations.AdminEmailPassword, password: migrations.AdminEmailPassword},
			wantErr: false,
		},
		{
			name:    "Invalid credentials admin",
			admin:   args{email: "invalid_" + migrations.AdminEmailPassword, password: "no_admin@admin.com"},
			wantErr: true,
		},
		{
			name:    "Valid credentials user",
			user:    args{email: migrations.UserEmailPassword, password: migrations.UserEmailPassword},
			wantErr: false,
		},
		{
			name:    "Invalid credentials user",
			user:    args{email: "invalid_" + migrations.UserEmailPassword, password: migrations.UserEmailPassword},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(defaultURL)
			if tt.admin.email != "" {
				c = NewClient(defaultURL, WithAdminEmailPassword(tt.admin.email, tt.admin.password))
			} else if tt.user.email != "" {
				c = NewClient(defaultURL, WithUserEmailPassword(tt.user.email, tt.user.password))
			}
			err := c.Authorize()
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestAuthorizeToken(t *testing.T) {
	tests := []struct {
		name       string
		validToken bool
		admin      bool
		user       bool
		wantErr    bool
	}{
		{
			name:       "Valid token admin",
			validToken: true,
			admin:      true,
			wantErr:    false,
		},
		{
			name:       "Invalid token admin",
			validToken: false,
			admin:      true,
			wantErr:    true,
		},
		{
			name:       "Valid token user",
			validToken: true,
			user:       true,
			wantErr:    false,
		},
		{
			name:       "Invalid token user",
			validToken: false,
			user:       true,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(defaultURL)
			if tt.admin {
				var token string
				if tt.validToken {
					c = NewClient(defaultURL,
						WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword),
					)
					_ = c.Authorize()
					token = c.AuthStore().Token()
				} else {
					token = "invalid_token"
				}
				c = NewClient(defaultURL, WithAdminToken(token))
			} else if tt.user {
				var token string
				if tt.validToken {
					c = NewClient(defaultURL,
						WithUserEmailPassword(migrations.UserEmailPassword, migrations.UserEmailPassword),
					)
					_ = c.Authorize()
					token = c.AuthStore().Token()
				} else {
					token = "invalid_token"
				}
				c = NewClient(defaultURL, WithUserToken(token))
			}
			err := c.Authorize()
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
