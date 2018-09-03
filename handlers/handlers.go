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
)

const dataEndpoint = `\/data$`
const localAuthority = `?type=geography`

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

func setStatusCode404(req *http.Request, w http.ResponseWriter, err error) {
	status := http.StatusNotFound
	log.ErrorCtx(req.Context(), err, log.Data{"setting-response-status": status})
	w.WriteHeader(status)
}

//GeographyHomepageRender ...
func GeographyHomepageRender(rend RenderClient) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		var page geographyHomepage.Page

		homePageLink := url + `/code-lists` + localAuthority

		resp, err := http.Get(homePageLink)
		if err != nil {
			log.Error(err, log.Data{"error getting data from the code-lists api http.Get(" + homePageLink + ") for GeographyHomepageRender returned ": err})
			log.Error(err, log.Data{"test": "error http.Get"})
			setStatusCode(req, w, err)
			return
		}
		if resp.StatusCode == 404 {
			log.Error(err, log.Data{"error getting data from the code-lists api http.Get(" + homePageLink + ") for GeographyHomepageRender returned ": resp.StatusCode})
			setStatusCode404(req, w, err)
			return
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err, log.Data{"error getting the .Body data from the code-lists api http.Get(" + homePageLink + ") for GeographyHomepageRender .Body returned ": resp.Body})
			setStatusCode(req, w, err)
			return
		}
		var codelistresults models.CodeListResults
		err = json.Unmarshal(b, &codelistresults)
		if err != nil {
			log.Error(err, log.Data{"error Unmarshaling the .Body data from the code-lists api http.Get(" + homePageLink + ") for GeographyHomepageRender returned ": err})
			setStatusCode(req, w, err)
			return
		}
		var codelist models.CodeList
		err = json.Unmarshal(b, &codelist)
		if err != nil {
			log.Error(err, log.Data{"error Unmarshaling the .Body data from the code-lists api http.Get(" + homePageLink + ") for GeographyHomepageRender returned ": err})
			setStatusCode(req, w, err)
			return
		}

		var geographyTypes []geographyHomepage.Items
		geographyTypesLabel := ""
		geographyTypesID := ""
		for i := range codelistresults.Items {
			log.Debug("for loop", log.Data{
				"geographyTypesTest": codelistresults.Items[i],
			})

			geographyTypesID = codelistresults.Items[i].Links.Self.ID

			resp, err := http.Get(codelistresults.Items[i].Links.Editions.Href)
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
			geographyTypesLabel = codelistresults.Items[0].Label
			geographyTypes = append(geographyTypes, geographyHomepage.Items{Label: geographyTypesLabel, ID: geographyTypesID})
			log.Debug("for loop", log.Data{
				"geographyTypesLabel": geographyTypesLabel,
			})
		}

		page.Data.Items = geographyTypes
		page.Metadata.Title = "Geography"

		templateJSON, err := json.Marshal(page)
		if err != nil {
			log.Error(err, log.Data{"error Marshaling the page data for GeographyHomepageRender returned ": err})
			setStatusCode(req, w, err)
			return
		}
		templateHTML, err := rend.Do("geography-homepage", templateJSON)
		if err != nil {
			log.Error(err, log.Data{"error rendering the geography-homepage data from GeographyHomepageRender returned ": err})
			setStatusCode(req, w, err)
			return
		}

		w.Write(templateHTML)
		return
	}
}
