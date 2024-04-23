package aptos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

// TODO: rename 'NodeClient' (vs IndexerClient) ?
// what looks best for `import aptos "github.com/aptoslabs/aptos-go-sdk"` then aptos.NewClient() ?
type RestClient struct {
	ChainId int

	client  http.Client
	baseUrl url.URL
}

func NewClient(baseUrl string) (rc *RestClient, err error) {
	rc = new(RestClient)
	tu, err := url.Parse(baseUrl)
	if err != nil {
		rc = nil
		return
	}
	rc.baseUrl = *tu
	rc.client.Timeout = 60 * time.Second
	return
}

// TODO: set HTTP header "x-aptos-client: aptos-go-sdk/{version}"

// TODO: define a struct and Unmarshal JSON onto it
func (rc *RestClient) Account(address AccountAddress, ledger_version ...int) (data map[string]any, err error) {
	au := rc.baseUrl
	au.Path = path.Join(au.Path, "accounts", address.String())
	if len(ledger_version) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.Itoa(ledger_version[0]))
		au.RawQuery = params.Encode()
	}
	response, err := rc.client.Get(au.String())
	if response.StatusCode != 200 {
		err = fmt.Errorf("http error: %s", response.Status)
		return
	}
	blob, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	response.Body.Close()
	err = json.Unmarshal(blob, &data)
	return
}

func (rc *RestClient) AccountResource(address AccountAddress, resourceType string, ledger_version ...int) (data map[string]any, err error) {
	au := rc.baseUrl
	// TODO: offer a list of known-good resourceType string constants
	au.Path = path.Join(au.Path, "accounts", address.String(), "resource", resourceType)
	if len(ledger_version) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.Itoa(ledger_version[0]))
		au.RawQuery = params.Encode()
	}
	response, err := rc.client.Get(au.String())
	if response.StatusCode >= 400 {
		err = fmt.Errorf("http error: %s", response.Status)
		return
	}
	blob, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	response.Body.Close()
	err = json.Unmarshal(blob, &data)
	return
}

func (rc *RestClient) AccountResources(address AccountAddress, ledger_version ...int) (data map[string]any, err error) {
	au := rc.baseUrl
	au.Path = path.Join(au.Path, "accounts", address.String(), "resources")
	if len(ledger_version) > 0 {
		params := url.Values{}
		params.Set("ledger_version", strconv.Itoa(ledger_version[0]))
		au.RawQuery = params.Encode()
	}
	response, err := rc.client.Get(au.String())
	if response.StatusCode >= 400 {
		err = fmt.Errorf("http error: %s", response.Status)
		return
	}
	blob, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	response.Body.Close()
	err = json.Unmarshal(blob, &data)
	return
}

func (rc *RestClient) Info() (data map[string]any, err error) {
	au := rc.baseUrl
	response, err := rc.client.Get(au.String())
	if response.StatusCode >= 400 {
		err = fmt.Errorf("http error: %s", response.Status)
		return
	}
	blob, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error getting response data, %w", err)
		return
	}
	response.Body.Close()
	err = json.Unmarshal(blob, &data)
	return
}
