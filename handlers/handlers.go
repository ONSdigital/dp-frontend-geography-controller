package handlers

import (
	"encoding/json"
	"net/http"

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

//HomepageRender ...
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
		for i := range codeListResults.Items {

			typesID := codeListResults.Items[i].Links.Self.ID
			editionsListResults, err := cli.GetEditionslistData(codeListResults.Items[i].Links.Editions.Href)
			if err != nil {
				log.ErrorCtx(ctx, errors.WithMessage(err, "error geting editions list from code-lists data"), nil)
				setStatusCode(req, w, err)
				return
			}

			if editionsListResults.Items[0].Label != "" {
				types = append(types, geographyHomepage.Items{
					Label: editionsListResults.Items[0].Label,
					ID:    typesID,
				})
			} else {
				log.ErrorCtx(ctx, errors.WithMessage(err, "editions label is undefined"), log.Data{"code_list_id": typesID})
			}
		}

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
