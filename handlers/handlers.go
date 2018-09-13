package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/ONSdigital/dp-frontend-geography-controller/models"
	"github.com/ONSdigital/dp-frontend-models/model/geographyHomepage"

	"github.com/ONSdigital/go-ns/clients/codelist"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
)

const dataEndpoint = `\/data$`

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
		var page geographyHomepage.Page

		codeListResults, err := cli.GetCodelistData()
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "error geting code-lists data for geography"), nil)
			setStatusCode(req, w, err)
			return
		}

		var types []geographyHomepage.Items
		var wg sync.WaitGroup
		var mutex = &sync.Mutex{}
		for _, v := range codeListResults.Items {
			wg.Add(1)
			go func(codeListResults models.CodeListResults, cli *codelist.Client, v models.CodeList) {
				defer wg.Done()
				typesID := v.Links.Self.ID
				editionsListResults, err := cli.GetEditionslistData(v.Links.Editions.Href)
				if err != nil {
					return
				}

				if len(editionsListResults.Items) > 0 && editionsListResults.Items[0].Label != "" {
					mutex.Lock()
					defer mutex.Unlock()
					types = append(types, geographyHomepage.Items{
						Label: editionsListResults.Items[0].Label,
						ID:    typesID,
					})
				}
				return
			}(codeListResults, cli, v)
		}
		wg.Wait()

		page.Data.Items = types
		page.Metadata.Title = "Geography"

		templateJSON, err := json.Marshal(page)
		if err != nil {
			log.ErrorCtx(ctx, errors.WithMessage(err, "error marshaling page data"), nil)
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
