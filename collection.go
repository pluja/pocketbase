package pocketbase

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/duke-git/lancet/v2/convertor"
)

type Collection[T any] struct {
	*Client
	Name               string
	BaseCollectionPath string
}

func CollectionSet[T any](client *Client, collection string) *Collection[T] {
	return &Collection[T]{
		Client:             client,
		Name:               collection,
		BaseCollectionPath: client.url + "/api/collections/" + url.QueryEscape(collection),
	}
}

func (c *Collection[T]) Create(body T) (ResponseCreate, error) {
	var response ResponseCreate

	if err := c.Client.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", c.Name).
		SetBody(body).
		SetResult(&response)

	resp, err := request.Post(c.url + "/api/collections/{collection}/records")
	if err != nil {
		return response, fmt.Errorf("[create] can't send update request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[create] pocketbase returned status: %d, msg: %s, body: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			fmt.Sprintf("%+v", body), // TODO remove that after debugging
			ErrInvalidResponse,
		)
	}

	return *resp.Result().(*ResponseCreate), nil
}

func (c *Collection[T]) Update(id string, body T) error {
	if err := c.Client.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", c.Name).
		SetBody(body)

	resp, err := request.Patch(c.url + "/api/collections/{collection}/records/" + id)
	if err != nil {
		return fmt.Errorf("[update] can't send update request to pocketbase, err %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("[update] pocketbase returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	return nil
}

func (c *Collection[T]) Delete(id string) error {
	if err := c.Client.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", c.Name).
		SetPathParam("id", id)

	resp, err := request.Delete(c.url + "/api/collections/{collection}/records/{id}")
	if err != nil {
		return fmt.Errorf("[delete] can't send delete request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[delete] pocketbase returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	return nil
}

func (c *Collection[T]) List(params ParamsList) (ResponseList[T], error) {
	var response ResponseList[T]

	if err := c.Client.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", c.Name)

	if params.Page > 0 {
		request.SetQueryParam("page", convertor.ToString(params.Page))
	}
	if params.Size > 0 {
		request.SetQueryParam("perPage", convertor.ToString(params.Size))
	}
	if params.Filters != "" {
		request.SetQueryParam("filter", params.Filters)
	}
	if params.Sort != "" {
		request.SetQueryParam("sort", params.Sort)
	}
	if params.Expand != "" {
		request.SetQueryParam("expand", params.Expand)
	}
	if params.Fields != "" {
		request.SetQueryParam("fields", params.Fields)
	}

	resp, err := request.Get(c.url + "/api/collections/{collection}/records")
	if err != nil {
		return response, fmt.Errorf("[list] can't send update request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[list] pocketbase returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	var responseRef any = &response
	if params.hackResponseRef != nil {
		responseRef = params.hackResponseRef
	}
	if err := json.Unmarshal(resp.Body(), responseRef); err != nil {
		return response, fmt.Errorf("[list] can't unmarshal response, err %w", err)
	}
	return response, nil
}

func (c *Collection[T]) FullList(params ParamsList) (ResponseList[T], error) {
	var response ResponseList[T]
	params.Page = 1
	params.Size = 500

	if err := c.Client.Authorize(); err != nil {
		return response, err
	}

	r, e := c.List(params)
	if e != nil {
		return response, e
	}
	response.Items = append(response.Items, r.Items...)
	response.Page = r.Page
	response.PerPage = r.PerPage
	response.TotalItems = r.TotalItems
	response.TotalPages = r.TotalPages

	for i := 2; i <= r.TotalPages; i++ { // Start from page 2 because first page is already fetched
		params.Page = i
		r, e := c.List(params)
		if e != nil {
			return response, e
		}
		response.Items = append(response.Items, r.Items...)
	}

	return response, nil
}

func (c *Collection[T]) One(id string) (T, error) {
	var response T

	err := c.OneTo(id, &response)

	return response, err
}

func (c *Collection[T]) OneTo(id string, result any) error {
	if err := c.Client.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", c.Name).
		SetPathParam("id", id)

	resp, err := request.Get(c.url + "/api/collections/{collection}/records/{id}")
	if err != nil {
		return fmt.Errorf("[oneTo] can't send get request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[oneTo] pocketbase returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), result); err != nil {
		return fmt.Errorf("[oneTo] can't unmarshal response, err %w", err)
	}

	return nil
}

// Get one record with params (only fields and expand supported)
func (c *Collection[T]) OneWithParams(id string, params ParamsList) (T, error) {
	var response T

	if err := c.Client.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", c.Name).
		SetPathParam("id", id).
		SetQueryParam("fields", params.Fields).
		SetQueryParam("expand", params.Expand)

	resp, err := request.Get(c.url + "/api/collections/{collection}/records/{id}")
	if err != nil {
		return response, fmt.Errorf("[one] can't send update request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[one] pocketbase returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[one] can't unmarshal response, err %w", err)
	}
	return response, nil
}
