package pocketbase

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pocketbase/pocketbase/core"
)

var ErrInvalidResponse = errors.New("invalid response")

type (
	Client struct {
		client     *resty.Client
		url        string
		authorizer authStore
		token      string
		sseDebug   bool
		restDebug  bool
	}
	ClientOption func(*Client)
)

func EnvIsTruthy(key string) bool {
	val := strings.ToLower(os.Getenv(key))
	return val == "1" || val == "true" || val == "yes"
}

func NewClient(url string, opts ...ClientOption) *Client {
	client := resty.New()
	client.
		SetRetryCount(3).
		SetRetryWaitTime(3 * time.Second).
		SetRetryMaxWaitTime(10 * time.Second)

	c := &Client{
		client:     client,
		url:        url,
		authorizer: authorizeNoOp{},
	}
	opts = append([]ClientOption{}, opts...)
	if EnvIsTruthy("REST_DEBUG") {
		opts = append(opts, WithRestDebug())
	}
	if EnvIsTruthy("SSE_DEBUG") {
		opts = append(opts, WithSseDebug())
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithRestDebug() ClientOption {
	return func(c *Client) {
		c.restDebug = true
		c.client.SetDebug(true)
	}
}

func WithSseDebug() ClientOption {
	return func(c *Client) {
		c.sseDebug = true
	}
}

func WithAdminEmailPassword22(email, password string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeEmailPassword(c.client, c.url+"/api/admins/auth-with-password", email, password)
	}
}

func WithAdminEmailPassword(email, password string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeEmailPassword(c.client, c.url+fmt.Sprintf("/api/collections/%s/auth-with-password", core.CollectionNameSuperusers), email, password)
	}
}

func WithUserEmailPassword(email, password string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeEmailPassword(c.client, c.url+"/api/collections/users/auth-with-password", email, password)
	}
}

func WithUserEmailPasswordAndCollection(email, password, collection string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeEmailPassword(c.client, c.url+"/api/collections/"+collection+"/auth-with-password", email, password)
	}
}

func WithAdminToken22(token string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeToken(c.client, c.url+"/api/admins/auth-refresh", token)
	}
}

func WithAdminToken(token string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeToken(c.client, c.url+fmt.Sprintf("/api/collections/%s/auth-refresh", core.CollectionNameSuperusers), token)
	}
}

func WithUserToken(token string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeToken(c.client, c.url+"/api/collections/users/auth-refresh", token)
	}
}

func (c *Client) Authorize() error {
	return c.authorizer.authorize()
}

func (c *Client) Get(path string, result any, onRequest func(*resty.Request), onResponse func(*resty.Response)) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json")
	if onRequest != nil {
		onRequest(request)
	}

	resp, err := request.Get(c.url + path)
	if err != nil {
		return fmt.Errorf("[get] can't send get request to pocketbase, err %w", err)
	}
	if onResponse != nil {
		onResponse(resp)
	}
	if resp.IsError() {
		return fmt.Errorf("[get] pocketbase returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), result); err != nil {
		return fmt.Errorf("[get] failed to unmarshal response: %w", err)
	}

	return nil
}

func (c *Client) Collection(name string) *Collection[map[string]any] {
	return &Collection[map[string]any]{
		Client: c,
		Name:   name,
	}
}

func (c *Client) AuthStore() authStore {
	return c.authorizer
}

func (c *Client) Backup() Backup {
	return Backup{
		Client: c,
	}
}

func (c *Client) Files() Files {
	return Files{
		Client: c,
	}
}
