package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-frontend-geography-controller/config"
	"github.com/ONSdigital/dp-frontend-models/model/geographyHomepage"

	"github.com/ONSdigital/dp-frontend-geography-controller/models"

	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
)

const dataEndpoint = `\/data$`

var url = config.Get().CodeListsAPIURL

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

//GeographyHomepageRender ...
func GeographyHomepageRender(rend RenderClient) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		var page geographyHomepage.Page
		ctx := req.Context()

		homePageLink := url + `/code-lists?type=geography`

		resp, err := http.Get(homePageLink)
		if err != nil {
			err = errors.Wrap(err, "error rendering homepage - failed to get data from the code-lists api")
			log.ErrorCtx(ctx, err, log.Data{"error": err, "url": homePageLink})
			setStatusCode(req, w, err)
			return
		}
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			err = errors.Wrap(err, "error reading the .Body data")
			log.ErrorCtx(ctx, err, log.Data{"error": err, "url": homePageLink})
			setStatusCode(req, w, err)
			return
		}
		var codelistresults models.CodeListResults
		err = json.Unmarshal(b, &codelistresults)
		if err != nil {
			err = errors.Wrap(err, "error unmarshaling from .CodeListResults")
			log.ErrorCtx(ctx, err, log.Data{"error": err, "url": homePageLink})
			setStatusCode(req, w, err)
			return
		}

		var geographyTypes []geographyHomepage.Items
		for i := range codelistresults.Items {

			geographyTypesID := codelistresults.Items[i].Links.Self.ID
			typesListLink := codelistresults.Items[i].Links.Editions.Href

			resp, err := http.Get(typesListLink)
			if err != nil {
				err = errors.Wrap(err, "error rendering geography types list - failed to get data from the code-lists api")
				log.ErrorCtx(ctx, err, log.Data{"error": err, "url": typesListLink})
				setStatusCode(req, w, err)
				return
			}
			defer resp.Body.Close()
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				err = errors.Wrap(err, "error reading the .Body data")
				log.ErrorCtx(ctx, err, log.Data{"error": err, "url": typesListLink})
				setStatusCode(req, w, err)
				return
			}
			var codelistresults models.CodeListResults
			err = json.Unmarshal(b, &codelistresults)
			if err != nil {
				err = errors.Wrap(err, "error unmarshaling from .CodeListResults")
				log.ErrorCtx(ctx, err, log.Data{"error": err, "url": typesListLink})
				setStatusCode(req, w, err)
				return
			}
			geographyTypesLabel := codelistresults.Items[0].Label
			geographyTypes = append(geographyTypes, geographyHomepage.Items{Label: geographyTypesLabel, ID: geographyTypesID})
		}

		page.Data.Items = geographyTypes
		page.Metadata.Title = "Geography"

		templateJSON, err := json.Marshal(page)
		if err != nil {
			err = errors.Wrap(err, "error marshaling page data")
			log.ErrorCtx(ctx, err, log.Data{"error": err})
			setStatusCode(req, w, err)
			return
		}
		templateHTML, err := rend.Do("geography-homepage", templateJSON)
		if err != nil {
			err = errors.Wrap(err, "error rendering homepage")
			log.ErrorCtx(ctx, err, log.Data{"error": err})
			setStatusCode(req, w, err)
			return
		}

		w.Write(templateHTML)
		return
	}
}
