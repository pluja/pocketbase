package pocketbase

import (
	"encoding/json"
	"fmt"
)

type otpResponse struct {
	Enabled  bool  `json:"enabled"`
	Duration int64 `json:"duration"` // in seconds
}

type mfaResponse struct {
	Enabled  bool  `json:"enabled"`
	Duration int64 `json:"duration"` // in seconds
}

type passwordResponse struct {
	IdentityFields []string `json:"identityFields"`
	Enabled        bool     `json:"enabled"`
}

type oauth2Response struct {
	Providers []providerInfo `json:"providers"`
	Enabled   bool           `json:"enabled"`
}

type providerInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	State       string `json:"state"`
	AuthURL     string `json:"authURL"`

	// @todo
	// deprecated: use AuthURL instead
	// AuthUrl will be removed after dropping v0.22 support
	AuthUrl string `json:"authUrl"`

	// technically could be omitted if the provider doesn't support PKCE,
	// but to avoid breaking existing typed clients we'll return them as empty string
	CodeVerifier        string `json:"codeVerifier"`
	CodeChallenge       string `json:"codeChallenge"`
	CodeChallengeMethod string `json:"codeChallengeMethod"`
}

// Borrowed from https://github.com/pocketbase/pocketbase/blob/844f18cac379fc749493dc4dd73638caa89167a1/apis/record_auth_methods.go#L52
type AuthMethodsResponse struct {
	Password passwordResponse `json:"password"`
	OAuth2   oauth2Response   `json:"oauth2"`
	MFA      mfaResponse      `json:"mfa"`
	OTP      otpResponse      `json:"otp"`

	// legacy fields
	// @todo remove after dropping v0.22 support
	AuthProviders    []providerInfo `json:"authProviders"`
	UsernamePassword bool           `json:"usernamePassword"`
	EmailPassword    bool           `json:"emailPassword"`
}

// ListAuthMethods returns all available collection auth methods.
func (c *Collection[T]) ListAuthMethods() (AuthMethodsResponse, error) {
	var response AuthMethodsResponse
	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json")

	resp, err := request.Get(c.BaseCollectionPath + "/auth-methods")
	if err != nil {
		return response, fmt.Errorf("[records] can't send ListAuthMethods request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[records] pocketbase returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[records] can't unmarshal response, err %w", err)
	}
	return response, nil
}

type (
	AuthWithPasswordResponse struct {
		Record Record `json:"record"`
		Token  string `json:"token"`
	}

	Record struct {
		Avatar          string `json:"avatar"`
		CollectionID    string `json:"collectionId"`
		CollectionName  string `json:"collectionName"`
		Created         string `json:"created"`
		Email           string `json:"email"`
		EmailVisibility bool   `json:"emailVisibility"`
		ID              string `json:"id"`
		Name            string `json:"name"`
		Updated         string `json:"updated"`
		Username        string `json:"username"`
		Verified        bool   `json:"verified"`
	}
)

// AuthWithPassword authenticate a single auth collection record via its username/email and password.
//
// On success, this method also automatically updates
// the client's AuthStore data and returns:
// - the authentication token via the AuthWithPasswordResponse
// - the authenticated record model
func (c *Collection[T]) AuthWithPassword(username string, password string) (AuthWithPasswordResponse, error) {
	var response AuthWithPasswordResponse
	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"identity": username,
			"password": password,
		})

	resp, err := request.Post(c.BaseCollectionPath + "/auth-with-password")
	if err != nil {
		return response, fmt.Errorf("[records] can't send auth-with-password request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[records] pocketbase returned status at auth-with-password: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[records] can't unmarshal auth-with-password-response, err %w", err)
	}

	c.token = response.Token
	return response, nil
}

type AuthWithOauth2Response struct {
	Token string `json:"token"`
}

// AuthWithOAuth2Code authenticate a single auth collection record with OAuth2 code.
//
// If you don't have an OAuth2 code you may also want to check `authWithOAuth2` method.
//
// On success, this method also automatically updates
// the client's AuthStore data and returns:
// - the authentication token via the model
// - the authenticated record model
// - the OAuth2 account data (eg. name, email, avatar, etc.)
func (c *Collection[T]) AuthWithOAuth2Code(provider string, code string, codeVerifier string, redirectURL string) (AuthWithOauth2Response, error) {
	var response AuthWithOauth2Response
	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"provider":     provider,
			"code":         code,
			"codeVerifier": codeVerifier,
			"redirectUrl":  redirectURL,
			//"createData":   createData,
		})

	resp, err := request.Post(c.BaseCollectionPath + "/auth-with-oauth2")
	if err != nil {
		return response, fmt.Errorf("[records] can't send auth-with-oauth2 request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[records] pocketbase returned status at auth-with-oauth2: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[records] can't unmarshal auth-with-oauth2-response, err %w", err)
	}

	c.token = response.Token
	return response, nil
}

type AuthRefreshResponse struct {
	Record struct {
		Avatar          string `json:"avatar"`
		CollectionID    string `json:"collectionId"`
		CollectionName  string `json:"collectionName"`
		Created         string `json:"created"`
		Email           string `json:"email"`
		EmailVisibility bool   `json:"emailVisibility"`
		ID              string `json:"id"`
		Name            string `json:"name"`
		Updated         string `json:"updated"`
		Username        string `json:"username"`
		Verified        bool   `json:"verified"`
	} `json:"record"`
	Token string `json:"token"`
}

// AuthRefresh refreshes the current authenticated record instance and
// * returns a new token and record data.
func (c *Collection[T]) AuthRefresh() (AuthRefreshResponse, error) {
	var response AuthRefreshResponse
	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(c.token)

	resp, err := request.Post(c.BaseCollectionPath + "/auth-refresh")
	if err != nil {
		return response, fmt.Errorf("[records] can't send auth-refresh request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[records] pocketbase returned status at auth-refresh: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[records] can't unmarshal auth-refresh-response, err %w", err)
	}

	c.token = response.Token
	return response, nil
}

// RequestVerification sends auth record verification email request.
func (c *Collection[T]) RequestVerification(email string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"email": email,
		})
	resp, err := request.Post(c.BaseCollectionPath + "/request-verification")
	if err != nil {
		return fmt.Errorf("[records] can't send request-verification request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at request-verification: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

// ConfirmVerification confirms auth record email verification request.
//
// If the current `client.authStore.model` matches with the auth record from the token,
// then on success the `client.authStore.model.verified` will be updated to `true`.
func (c *Collection[T]) ConfirmVerification(verificationToken string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"token": verificationToken,
		})
	resp, err := request.Post(c.BaseCollectionPath + "/confirm-verification")
	if err != nil {
		return fmt.Errorf("[records] can't send confirm-verification request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at confirm-verification: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

// RequestPasswordReset sends auth record password reset request
func (c *Collection[T]) RequestPasswordReset(email string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"email": email,
		})
	resp, err := request.Post(c.BaseCollectionPath + "/request-password-reset")
	if err != nil {
		return fmt.Errorf("[records] can't send request-password-reset request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at request-password-reset: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

// ConfirmPasswordReset confirms auth record password reset request.
func (c *Collection[T]) ConfirmPasswordReset(passwordResetToken string, password string, passwordConfirm string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"token":           passwordResetToken,
			"password":        password,
			"passwordConfirm": passwordConfirm,
		})
	resp, err := request.Post(c.BaseCollectionPath + "/confirm-password-reset")
	if err != nil {
		return fmt.Errorf("[records] can't send confirm-password-reset request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at confirm-password-reset: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

// RequestEmailChange sends an email change request to the authenticated record model.
func (c *Collection[T]) RequestEmailChange(newEmail string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"newEmail": newEmail,
		}).
		SetAuthToken(c.token)

	resp, err := request.Post(c.BaseCollectionPath + "/request-email-change")
	if err != nil {
		return fmt.Errorf("[records] can't send request-email-change request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at request-email-change: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

// ConfirmEmailChange confirms auth record's new email address.
func (c *Collection[T]) ConfirmEmailChange(emailChangeToken string, password string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"token":    emailChangeToken,
			"password": password,
		}).
		SetAuthToken(c.token)

	resp, err := request.Post(c.BaseCollectionPath + "/confirm-email-change")
	if err != nil {
		return fmt.Errorf("[records] can't send confirm-email-change request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at confirm-email-change: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

func (c *Collection[T]) baseCollectionPath() string {
	return c.BaseCollectionPath
}

func (c *Collection[T]) baseCrudPath() string {
	return c.BaseCollectionPath + "/records/"
}
