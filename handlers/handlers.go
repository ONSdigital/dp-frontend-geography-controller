package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"sync"

	"github.com/ONSdigital/dp-api-clients-go/headers"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/geography/area"
	"github.com/ONSdigital/dp-frontend-models/model/geography/homepage"
	"github.com/ONSdigital/dp-frontend-models/model/geography/list"
	"github.com/ONSdigital/go-ns/common"
	"github.com/gorilla/mux"

	"github.com/ONSdigital/dp-api-clients-go/codelist"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/log.go/log"
)

//go:generate moq -out mocks_handlers.go . CodeListClient RenderClient DatasetClient

// CodeListClient is an interface with methods required for a code-list client
type CodeListClient interface {
	GetGeographyCodeLists(ctx context.Context, userAuthToken string, serviceAuthToken string) (editions codelist.CodeListResults, err error)
	GetCodeListEditions(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string) (editions codelist.EditionsListResults, err error)
	GetCodes(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string) (codes codelist.CodesResults, err error)
	GetCodeByID(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (code codelist.CodeResult, err error)
	GetDatasetsByCode(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (datasets codelist.DatasetsResult, err error)
}

// DatasetClient is an interface with methods required for a dataset client
type DatasetClient interface {
	Get(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID string) (m dataset.DatasetDetails, err error)
}

// RenderClient is an interface with methods for require for rendering a template
type RenderClient interface {
	Do(string, []byte) ([]byte, error)
}

// ClientError is an interface that can be used to retrieve the status code if a client has errored
type ClientError interface {
	error
	Code() int
}

func setStatusCode(req *http.Request, w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if err, ok := err.(ClientError); ok {
		if err.Code() == http.StatusNotFound {
			status = err.Code()
		}
		log.Event(req.Context(), "setting response status", log.ERROR, log.Error(err), log.Data{"status": status})
	}
	w.WriteHeader(status)
}

//HomepageRender gets geography data from the code-list-api and formats for rendering
func HomepageRender(rend RenderClient, cli CodeListClient, enableLoop11 bool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		var page homepage.Page

		serviceAuthToken := getServiceAuthToken(ctx, req)
		userAuthToken := getUserAuthToken(ctx, req)

		codeListResults, err := cli.GetGeographyCodeLists(ctx, userAuthToken, serviceAuthToken)
		if err != nil {
			log.Event(ctx, "error getting geography code-lists", log.ERROR, log.Error(err))
			setStatusCode(req, w, err)
			return
		}

		var types []homepage.Item
		var wg sync.WaitGroup
		var mutex = &sync.Mutex{}
		for _, v := range codeListResults.Items {
			wg.Add(1)
			go func(codeListResults codelist.CodeListResults, cli CodeListClient, v codelist.CodeList) {
				defer wg.Done()
				typesID := v.Links.Self.ID
				editionsListResults, err := cli.GetCodeListEditions(ctx, userAuthToken, serviceAuthToken, typesID)
				if err != nil {
					log.Event(ctx, "Error doing GET editions for code-list", log.ERROR, log.Error(err), log.Data{
						"codeListID": typesID,
					})
					return
				}

				if len(editionsListResults.Items) > 0 && editionsListResults.Items[0].Label != "" {
					mutex.Lock()
					defer mutex.Unlock()
					types = append(types, homepage.Item{
						Label: editionsListResults.Items[0].Label,
						ID:    typesID,
						URI:   fmt.Sprintf("/geography/%s", typesID),
					})
				}
				return
			}(codeListResults, cli, v)
		}
		wg.Wait()

		sort.Slice(types, func(i, j int) bool {
			return types[i].Label < types[j].Label
		})

		page.Data.Items = types
		page.BetaBannerEnabled = true
		page.Metadata.Title = "Geography"
		page.EnableLoop11 = enableLoop11
		page.Breadcrumb = []model.TaxonomyNode{
			model.TaxonomyNode{
				Title: "Home",
				URI:   "https://www.ons.gov.uk",
			},
			model.TaxonomyNode{
				Title: "Geography",
				URI:   "/geography",
			},
		}

		templateJSON, err := json.Marshal(page)
		if err != nil {
			log.Event(ctx, "error marshaling geography code-lists page data", log.ERROR, log.Error(err))
			setStatusCode(req, w, err)
			return
		}
		templateHTML, err := rend.Do("geography-homepage", templateJSON)
		if err != nil {
			log.Event(ctx, "error rendering homepage", log.ERROR, log.Error(err))
			setStatusCode(req, w, err)
			return
		}

		w.Write(templateHTML)
		return
	}
}

//ListPageRender renders a list of codes associated to the first edition of a code-list
func ListPageRender(rend RenderClient, cli CodeListClient, enableLoop11 bool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		vars := mux.Vars(req)
		codeListID := vars["codeListID"]
		logData := log.Data{
			codeListID: codeListID,
		}
		var page list.Page
		serviceAuthToken := getServiceAuthToken(ctx, req)
		userAuthToken := getUserAuthToken(ctx, req)

		codeListEditions, err := cli.GetCodeListEditions(ctx, userAuthToken, serviceAuthToken, codeListID)
		if err != nil {
			log.Event(ctx, "error getting editions for a code-list", log.ERROR, log.Error(err), logData)
			setStatusCode(req, w, err)
			return
		}

		if codeListEditions.Count > 0 {
			edition := codeListEditions.Items[0]
			page.Metadata.Title = edition.Label

			log.Event(ctx, "getting codes for edition of a code list", log.Data{"edition": edition})
			codes, err := cli.GetCodes(ctx, userAuthToken, serviceAuthToken, codeListID, edition.Edition)
			if err != nil {
				logData["edition"] = edition.Edition
				log.Event(ctx, "error getting codes for an edition of a code-list", log.ERROR, log.Error(err), logData)
				setStatusCode(req, w, err)
				return
			}

			if codes.Count > 0 {
				var pageCodes []list.Item
				for _, item := range codes.Items {
					pageCodes = append(pageCodes, list.Item{
						ID:    item.ID,
						Label: item.Label,
						URI:   fmt.Sprintf("/geography/%s/%s", codeListID, item.ID),
					})
				}
				sort.Slice(pageCodes[:], func(i, j int) bool {
					return pageCodes[i].Label < pageCodes[j].Label
				})

				page.Data.Items = pageCodes
			}
		}
		page.BetaBannerEnabled = true
		page.EnableLoop11 = enableLoop11
		page.Breadcrumb = []model.TaxonomyNode{
			model.TaxonomyNode{
				Title: "Home",
				URI:   "https://www.ons.gov.uk",
			},
			model.TaxonomyNode{
				Title: "Geography",
				URI:   "/geography",
			},
			model.TaxonomyNode{
				Title: page.Metadata.Title,
				URI:   fmt.Sprintf("/geography/%s", codeListID),
			},
		}

		templateJSON, err := json.Marshal(page)
		if err != nil {
			log.Event(ctx, "error marshalling geography list page data to JSON", log.ERROR, log.Error(err), logData)
			setStatusCode(req, w, err)
			return
		}
		templateHTML, err := rend.Do("geography-list", templateJSON)
		if err != nil {
			log.Event(ctx, "error getting HTML of list of geographic areas", log.ERROR, log.Error(err), logData)
			setStatusCode(req, w, err)
			return
		}

		w.Write(templateHTML)
		return
	}
}

//AreaPageRender gets data about a specific code, get what datasets are associated with the code and get information
// about those datasets, maps it and passes it to the renderer
func AreaPageRender(rend RenderClient, cli CodeListClient, dcli DatasetClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		vars := mux.Vars(req)
		codeListID := vars["codeListID"]
		codeID := vars["codeID"]
		collectionID := vars["codeID"]

		logData := log.Data{
			codeListID: codeListID,
			codeID:     codeID,
		}

		var page area.Page
		serviceAuthToken := getServiceAuthToken(ctx, req)
		userAuthToken := getUserAuthToken(ctx, req)

		codeListEditions, err := cli.GetCodeListEditions(ctx, userAuthToken, serviceAuthToken, codeListID)
		if err != nil {
			log.Event(ctx, "error getting editions for a code-list", log.ERROR, log.Error(err), logData)
			setStatusCode(req, w, err)
			return
		}

		var parentName string

		if codeListEditions.Count > 0 {
			edition := codeListEditions.Items[0]
			parentName = edition.Label

			log.Event(ctx, "getting data about code", log.INFO, log.Data{"edition": edition})
			codeData, err := cli.GetCodeByID(ctx, userAuthToken, serviceAuthToken, codeListID, edition.Edition, codeID)
			if err != nil {
				log.Event(ctx, "error getting code data", log.ERROR, log.Error(err), logData)
				setStatusCode(req, w, err)
				return
			}
			page.Metadata.Title = codeData.Label

			datasetsResp, err := cli.GetDatasetsByCode(ctx, userAuthToken, serviceAuthToken, codeListID, edition.Edition, codeID)
			if err != nil {
				log.Event(ctx, "error getting datasets related to code", log.ERROR, log.Error(err), logData)
				setStatusCode(req, w, err)
				return
			}

			if datasetsResp.Count > 0 {
				var datasets []area.Dataset
				var wg sync.WaitGroup
				var mutex = &sync.Mutex{}
				var gotErr bool
				for _, datasetResp := range datasetsResp.Datasets {
					wg.Add(1)
					go func(ctx context.Context, dcli DatasetClient, datasetResp codelist.Dataset) {
						defer wg.Done()
						datasetDetails, err := dcli.Get(ctx, userAuthToken, serviceAuthToken, collectionID, datasetResp.Links.Self.ID)
						if err != nil {
							gotErr = true
							log.Event(ctx, "error getting dataset", log.ERROR, log.Error(err), logData)
							return
						}
						datasetWebsiteURL, err := url.Parse(datasetResp.Editions[0].Links.LatestVersion.Href)
						if err != nil {
							gotErr = true
							log.Event(ctx, "error parsing dataset href", log.ERROR, log.Error(err), logData)
							return
						}
						mutex.Lock()
						defer mutex.Unlock()
						datasets = append(datasets, area.Dataset{
							ID:          datasetResp.Editions[0].Links.Self.ID,
							Label:       datasetDetails.Title,
							Description: datasetDetails.Description,
							URI:         datasetWebsiteURL.Path,
						})

						return
					}(ctx, dcli, datasetResp)
				}
				wg.Wait()
				if gotErr {
					setStatusCode(req, w, err)
					return
				}
				page.Data.Datasets = datasets
			}
		}

		page.Data.Attributes.Code = codeID
		page.BetaBannerEnabled = true
		page.Breadcrumb = getAreaPageRenderBreadcrumb(parentName, page.Metadata.Title, codeListID, codeID)

		templateJSON, err := json.Marshal(page)
		if err != nil {
			log.Event(ctx, "error marshalling geography area page data to JSON", log.ERROR, log.Error(err), logData)
			setStatusCode(req, w, err)
			return
		}
		templateHTML, err := rend.Do("geography-area", templateJSON)
		if err != nil {
			log.Event(ctx, "error getting HTML of geographic area page", log.ERROR, log.Error(err), logData)
			setStatusCode(req, w, err)
			return
		}

		w.Write(templateHTML)
		return
	}
}

func getAreaPageRenderBreadcrumb(parentName string, pageTitle string, codeListID string, codeID string) []model.TaxonomyNode {
	return []model.TaxonomyNode{
		model.TaxonomyNode{
			Title: "Home",
			URI:   "https://www.ons.gov.uk",
		},
		model.TaxonomyNode{
			Title: "Geography",
			URI:   "/geography",
		},
		model.TaxonomyNode{
			Title: parentName,
			URI:   fmt.Sprintf("/geography/%s", codeListID),
		},
		model.TaxonomyNode{
			Title: pageTitle,
			URI:   fmt.Sprintf("/geography/%s/%s", codeListID, codeID),
		},
	}
}

func getUserAuthToken(ctx context.Context, req *http.Request) string {
	token, err := headers.GetUserAuthToken(req)
	if err == nil {
		return token
	}

	cookie, err := req.Cookie(common.FlorenceCookieKey)
	if err != nil && err == http.ErrNoCookie {
		return ""
	} else if err != nil {
		log.Event(ctx, "error getting access token cookie from request", log.ERROR, log.Error(err))
		return ""
	}
	return cookie.Value
}

func getServiceAuthToken(ctx context.Context, req *http.Request) string {
	token, _ := headers.GetServiceAuthToken(req)
	return token
}
