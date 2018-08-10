package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-frontend-models/model/geographyHomepage"

	"github.com/ONSdigital/dp-frontend-geography-controller/models"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/log"
)

const dataEndpoint = `\/data$`
const localAuthority = `?type=geography`

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

func forwardFlorenceTokenIfRequired(req *http.Request) *http.Request {
	if len(req.Header.Get(common.FlorenceHeaderKey)) > 0 {
		ctx := common.SetFlorenceIdentity(req.Context(), req.Header.Get(common.FlorenceHeaderKey))
		return req.WithContext(ctx)
	}
	return req
}

//GeographyRender ...
func GeographyRender(rend RenderClient) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		var page geographyHomepage.Page

		resp, err := http.Get(`https://api.dev.cmd.onsdigital.co.uk/v1/code-lists` + localAuthority)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}
		var codelistresults models.CodeListResults
		err = json.Unmarshal(b, &codelistresults)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}
		var codelist models.CodeList
		err = json.Unmarshal(b, &codelist)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}

		geographyTypes := ""
		for i := range codelistresults.Items {
			log.Debug("test", log.Data{
				"geographyTypesTest": codelistresults.Items[i],
			})

			geographyTypes = geographyTypes + codelistresults.Items[i].Name

		}

		page.Data.AreaTypes = []geographyHomepage.AreaType{
			{Name: "Countries"},
			{Name: "Regions"},
			{Name: geographyTypes},
		}

		page.Metadata.Title = "Geography"

		templateJSON, err := json.Marshal(page)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}
		templateHTML, err := rend.Do("geography-homepage", templateJSON)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}

		w.Write(templateHTML)
		return
	}
}
