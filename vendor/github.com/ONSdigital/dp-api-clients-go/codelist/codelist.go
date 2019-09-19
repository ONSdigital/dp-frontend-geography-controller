package codelist

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/log.go/log"
)

const service = "code-list-api"

var _ error = ErrInvalidCodelistAPIResponse{}

// Client is a codelist api client which can be used to make requests to the server
type Client struct {
	cli rchttp.Clienter
	url string
}

// ErrInvalidCodelistAPIResponse is returned when the codelist api does not respond
// with a valid status
type ErrInvalidCodelistAPIResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidCodelistAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from codelist api - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

// Code returns the status code received from code list api if an error is returned
func (e ErrInvalidCodelistAPIResponse) Code() int {
	return e.actualCode
}

// New creates a new instance of Client with a given filter api url
func New(codelistAPIURL string) *Client {
	return &Client{
		cli: rchttp.NewClient(),
		url: codelistAPIURL,
	}
}

// Healthcheck calls the healthcheck endpoint on the api and alerts the caller of any errors
func (c *Client) Healthcheck() (string, error) {
	resp, err := c.cli.Get(context.Background(), c.url+"/healthcheck")
	if err != nil {
		return service, err
	}

	if resp.StatusCode != http.StatusOK {
		return service, &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, "/healthcheck"}
	}

	return service, nil
}

// GetValues returns dimension values from the codelist api
func (c *Client) GetValues(ctx context.Context, userAuthToken string, serviceAuthToken string, id string) (DimensionValues, error) {
	var vals DimensionValues
	uri := fmt.Sprintf("%s/code-lists/%s/codes", c.url, id)

	log.Event(ctx, "retrieving codes from codelist", log.Data{
		"method":  "GET",
		"uri":     uri,
		"service": service,
	})

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, uri)
	if err != nil {
		return vals, err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return vals, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return vals, err
	}

	err = json.Unmarshal(b, &vals)
	return vals, err
}

// GetIDNameMap returns dimension values in the form of an id name map
func (c *Client) GetIDNameMap(ctx context.Context, userAuthToken string, serviceAuthToken string, id string) (map[string]string, error) {
	uri := fmt.Sprintf("%s/code-lists/%s/codes", c.url, id)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, uri)
	if err != nil {
		return nil, err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return nil, &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var vals DimensionValues
	if err = json.Unmarshal(body, &vals); err != nil {
		return nil, err
	}

	idNames := make(map[string]string)
	for _, val := range vals.Items {
		idNames[val.ID] = val.Label
	}

	return idNames, nil
}

//GetGeographyCodeLists returns the geography codelists
func (c *Client) GetGeographyCodeLists(ctx context.Context, userAuthToken string, serviceAuthToken string) (CodeListResults, error) {
	uri := fmt.Sprintf("%s/code-lists?type=geography", c.url)
	var results CodeListResults

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, uri)
	if err != nil {
		return results, err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return results, &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return results, err
	}

	err = json.Unmarshal(b, &results)
	if err != nil {
		return results, err
	}
	return results, nil
}

//GetCodeListEditions returns the editions for a codelist
func (c *Client) GetCodeListEditions(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string) (EditionsListResults, error) {
	url := fmt.Sprintf("%s/code-lists/%s/editions", c.url, codeListID)
	var editionsList EditionsListResults

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, url)
	if err != nil {
		return editionsList, err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != 200 {
		return editionsList, &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, url}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return editionsList, err
	}

	err = json.Unmarshal(b, &editionsList)
	if err != nil {
		return editionsList, err
	}

	return editionsList, nil
}

//GetCodes returns the codes for a specific edition of a code list
func (c *Client) GetCodes(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string) (CodesResults, error) {
	var codes CodesResults
	url := fmt.Sprintf("%s/code-lists/%s/editions/%s/codes", c.url, codeListID, edition)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, url)
	if err != nil {
		return codes, err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return codes, &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, url}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return codes, err
	}

	err = json.Unmarshal(b, &codes)
	if err != nil {
		return codes, err
	}

	return codes, nil
}

// GetCodeByID returns information about a code
func (c *Client) GetCodeByID(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (CodeResult, error) {
	var code CodeResult
	url := fmt.Sprintf("%s/code-lists/%s/editions/%s/codes/%s", c.url, codeListID, edition, codeID)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, url)
	if err != nil {
		return code, err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return code, &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, url}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return code, err
	}

	err = json.Unmarshal(b, &code)
	if err != nil {
		return code, err
	}

	return code, nil
}

// GetDatasetsByCode returns datasets containing the codelist codeID.
func (c *Client) GetDatasetsByCode(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (DatasetsResult, error) {
	var datasets DatasetsResult
	url := fmt.Sprintf("%s/code-lists/%s/editions/%s/codes/%s/datasets", c.url, codeListID, edition, codeID)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, url)
	if err != nil {
		return datasets, err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return datasets, &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, url}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return datasets, err
	}

	err = json.Unmarshal(b, &datasets)
	if err != nil {
		return datasets, err
	}
	return datasets, nil
}

// doGetWithAuthHeaders executes clienter.Do setting the service authentication token as a request header. Returns the http.Response and any error.
// It is the callers responsibility to ensure response.Body is closed on completion.
func (c *Client) doGetWithAuthHeaders(ctx context.Context, userAuthToken string, serviceAuthToken string, uri string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	common.AddFlorenceHeader(req, userAuthToken)
	common.AddServiceTokenHeader(req, serviceAuthToken)
	return c.cli.Do(ctx, req)
}

func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body == nil {
		return
	}

	if err := resp.Body.Close(); err != nil {
		log.Event(ctx, "error closing http response body", log.Error(err))
	}
}