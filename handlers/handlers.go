package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"sync"

	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/geography/area"
	"github.com/ONSdigital/dp-frontend-models/model/geography/homepage"
	"github.com/ONSdigital/dp-frontend-models/model/geography/list"
	"github.com/gorilla/mux"

	"github.com/ONSdigital/go-ns/clients/codelist"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
)

// CodeListClient is an interface with methods required for a code-list client
type CodeListClient interface {
	healthcheck.Client
	GetCodeListEditions(codeListID string) (editions codelist.EditionsListResults, err error)
	GetCodes(codeListID string, edition string) (codes codelist.CodesResults, err error)
	GetCodeByID(codeListID string, edition string, codeID string) (code codelist.CodeResult, err error)
	GetDatasetsByCode(codeListID string, edition string, codeID string) (datasets codelist.DatasetsResult, err error)
}

// DatasetClient is an interface with methods required for a dataset client
type DatasetClient interface {
	healthcheck.Client
	Get(ctx context.Context, id string) (m dataset.Model, err error)
}

// RenderClient is an interface with methods for require for rendering a template
type RenderClient interface {
	healthcheck.Client
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
	}
	log.ErrorCtx(req.Context(), err, log.Data{"setting-response-status": status})
	w.WriteHeader(status)
}

//HomepageRender gets geography data from the code-list-api and formats for rendering
func HomepageRender(rend RenderClient, cli *codelist.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		var page homepage.Page

		codeListResults, err := cli.GetGeographyCodeLists()
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "error getting geography code-lists"), nil)
			setStatusCode(req, w, err)
			return
		}

		var types []homepage.Item
		var wg sync.WaitGroup
		var mutex = &sync.Mutex{}
		for _, v := range codeListResults.Items {
			wg.Add(1)
			go func(codeListResults codelist.CodeListResults, cli *codelist.Client, v codelist.CodeList) {
				defer wg.Done()
				typesID := v.Links.Self.ID
				editionsListResults, err := cli.GetCodeListEditions(typesID)
				if err != nil {
					log.ErrorCtx(ctx, errors.WithMessage(err, "Error doing GET editions for code-list"), log.Data{
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

		page.Data.Items = types
		page.Metadata.Title = "Geography"
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
			log.ErrorCtx(ctx, errors.WithMessage(err, "error marshaling geography code-lists page data"), nil)
			setStatusCode(req, w, err)
			return
		}
		templateHTML, err := rend.Do("geography-homepage", templateJSON)
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "error rendering homepage"), nil)
			setStatusCode(req, w, err)
			return
		}

		w.Write(templateHTML)
		return
	}
}

//ListPageRender renders a list of codes associated to the first edition of a code-list
func ListPageRender(rend RenderClient, cli CodeListClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		vars := mux.Vars(req)
		codeListID := vars["codeListID"]
		logData := log.Data{
			codeListID: codeListID,
		}
		var page list.Page

		codeListEditions, err := cli.GetCodeListEditions(codeListID)
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "error getting editions for a code-list"), logData)
			setStatusCode(req, w, err)
			return
		}

		if codeListEditions.Count > 0 {
			edition := codeListEditions.Items[0]
			page.Metadata.Title = edition.Label

			log.InfoCtx(ctx, "getting codes for edition of a code list", log.Data{"edition": edition})
			codes, err := cli.GetCodes(codeListID, edition.Edition)
			if err != nil {
				logData["edition"] = edition.Edition
				log.ErrorCtx(ctx, errors.WithMessage(err, "error getting codes for an edition of a code-list"), logData)
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
			log.ErrorCtx(ctx, errors.WithMessage(err, "error marshalling geography list page data to JSON"), logData)
			setStatusCode(req, w, err)
			return
		}
		templateHTML, err := rend.Do("geography-list", templateJSON)
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "error getting HTML of list of geographic areas"), logData)
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
		logData := log.Data{
			codeListID: codeListID,
			codeID:     codeID,
		}
		var page area.Page

		codeListEditions, err := cli.GetCodeListEditions(codeListID)
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "error getting editions for a code-list"), logData)
			setStatusCode(req, w, err)
			return
		}

		var parentName string

		if codeListEditions.Count > 0 {
			edition := codeListEditions.Items[0]
			parentName = edition.Label

			log.InfoCtx(ctx, "getting data about code", log.Data{"edition": edition})
			codeData, err := cli.GetCodeByID(codeListID, edition.Edition, codeID)
			if err != nil {
				log.ErrorCtx(ctx, errors.WithMessage(err, "error getting code data"), logData)
				setStatusCode(req, w, err)
				return
			}
			page.Metadata.Title = codeData.Label

			datasetsResp, err := cli.GetDatasetsByCode(codeListID, edition.Edition, codeID)
			if err != nil {
				log.ErrorCtx(ctx, errors.WithMessage(err, "error getting datasets related to code"), logData)
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
						datasetDetails, err := dcli.Get(ctx, datasetResp.Links.Self.ID)
						if err != nil {
							gotErr = true
							log.ErrorCtx(ctx, errors.WithMessage(err, "error getting dataset"), logData)
							return
						}
						datasetWebsiteURL, err := url.Parse(datasetResp.Editions[0].Links.LatestVersion.Href)
						if err != nil {
							gotErr = true
							log.ErrorCtx(ctx, errors.WithMessage(err, "error parsing dataset href"), logData)
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
				Title: parentName,
				URI:   fmt.Sprintf("/geography/%s", codeListID),
			},
			model.TaxonomyNode{
				Title: page.Metadata.Title,
				URI:   fmt.Sprintf("/geography/%s/%s", codeListID, codeID),
			},
		}

		templateJSON, err := json.Marshal(page)
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "error marshalling geography area page data to JSON"), logData)
			setStatusCode(req, w, err)
			return
		}
		templateHTML, err := rend.Do("geography-area", templateJSON)
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "error getting HTML of geographic area page"), logData)
			setStatusCode(req, w, err)
			return
		}

		w.Write(templateHTML)
		return
	}
}
