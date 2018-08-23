package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-frontend-models/model/geographyAreaPage"
	"github.com/ONSdigital/dp-frontend-models/model/geographyHomepage"
	"github.com/ONSdigital/dp-frontend-models/model/geographyListPage"

	"github.com/ONSdigital/dp-frontend-geography-controller/models"

	"github.com/gorilla/mux"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/log"
)

const dataEndpoint = `\/data$`
const typeGeography = `?type=geography`

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

func forwardFlorenceTokenIfRequired(req *http.Request) *http.Request {
	if len(req.Header.Get(common.FlorenceHeaderKey)) > 0 {
		ctx := common.SetFlorenceIdentity(req.Context(), req.Header.Get(common.FlorenceHeaderKey))
		return req.WithContext(ctx)
	}
	return req
}

//GeographyHomepageRender ...
func GeographyHomepageRender(rend RenderClient) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		var page geographyHomepage.Page
		homePageLink := `https://api.dev.cmd.onsdigital.co.uk/v1/code-lists` + typeGeography

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

		var geographyTypes []geographyHomepage.AreaType
		geographyTypesLabel := ""
		geographyTypesID := ""
		for i := range codelistresults.Items {
			log.Debug("test", log.Data{
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

			geographyTypes = append(geographyTypes, geographyHomepage.AreaType{Label: geographyTypesLabel, ID: geographyTypesID})

		}

		page.Data.AreaTypes = geographyTypes
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

//GeographyListpageRender ...
func GeographyListpageRender(rend RenderClient) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		var page geographyListPage.Page

		vars := mux.Vars(req)
		areaTypeID := vars["areaTypeID"]
		listPageLink := `https://api.dev.cmd.onsdigital.co.uk/v1/code-lists/` + areaTypeID + `/editions/2016/codes`

		resp, err := http.Get(listPageLink)
		if err != nil {
			log.Error(err, log.Data{"error getting data from the code-lists api http.Get(" + listPageLink + ") for GeographyListpageRender returned ": err})
			setStatusCode(req, w, err)
			return
		}
		if resp.StatusCode == 404 {
			log.Error(err, log.Data{"error getting data from the code-lists api http.Get(" + listPageLink + ") for GeographyListpageRender returned ": resp.StatusCode})
			setStatusCode404(req, w, err)
			return
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err, log.Data{"error getting the .Body data from the code-lists api http.Get(" + listPageLink + ") for GeographyListpageRender .Body returned ": resp.Body})
			setStatusCode(req, w, err)
			return
		}

		var codelistresults models.CodeListResults
		err = json.Unmarshal(b, &codelistresults)
		if err != nil {
			log.Error(err, log.Data{"error Unmarshaling the .Body data from the code-lists api http.Get(" + listPageLink + ") for GeographyListpageRender returned ": err})
			setStatusCode(req, w, err)
			return
		}
		var codelist models.CodeList
		err = json.Unmarshal(b, &codelist)
		if err != nil {
			log.Error(err, log.Data{"error Unmarshaling the .Body data from the code-lists api http.Get(" + listPageLink + ") for GeographyListpageRender returned ": err})
			setStatusCode(req, w, err)
			return
		}

		var geographyTypes []geographyListPage.AreaType
		for i := range codelistresults.Items {
			geographyTypes = append(geographyTypes, geographyListPage.AreaType{Label: codelistresults.Items[i].Label, ID: codelistresults.Items[i].ID})
		}

		page.Data.AreaTypes = geographyTypes
		page.Metadata.Title = areaTypeID

		templateJSON, err := json.Marshal(page)
		if err != nil {
			log.Error(err, log.Data{"error Marshaling the page data for GeographyListpageRender returned ": err})
			setStatusCode(req, w, err)
			return
		}
		templateHTML, err := rend.Do("geography-list-page", templateJSON)
		if err != nil {
			log.Error(err, log.Data{"error rendering the geography-list-page data from GeographyListpageRender returned ": err})
			setStatusCode(req, w, err)
			return
		}

		w.Write(templateHTML)
		return
	}
}

//GeographyAreapageRender ...
func GeographyAreapageRender(rend RenderClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var page geographyAreaPage.Page

		vars := mux.Vars(req)
		areaTypeID := vars["areaTypeID"]
		datasetLabel := vars["datasetLabel"]
		datasetID := vars["datasetID"]
		AreaPageLink := `https://api.dev.cmd.onsdigital.co.uk/v1/code-lists/` + areaTypeID + `/editions/2016/codes/` + datasetID + `/datasets`

		resp, err := http.Get(AreaPageLink)
		if err != nil {
			log.Error(err, log.Data{"error getting data from the code-lists api http.Get(" + AreaPageLink + ") for GeographyAreapageRender returned ": err})
			setStatusCode(req, w, err)
			return
		}
		if resp.StatusCode == 404 {
			log.Error(err, log.Data{"error getting data from the code-lists api http.Get(" + AreaPageLink + ") for GeographyAreapageRender returned ": resp.StatusCode})
			setStatusCode404(req, w, err)
			return
		}
		if err == nil {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error(err, log.Data{"error getting the .Body data from the code-lists api http.Get(" + AreaPageLink + ") for GeographyAreapageRender .Body returned ": resp.Body})
				setStatusCode(req, w, err)
				return
			}

			var datasetlistresults models.DatasetListResults
			err = json.Unmarshal(b, &datasetlistresults)
			if err != nil {
				log.Error(err, log.Data{"error Unmarshaling the .Body data from the code-lists api http.Get(" + AreaPageLink + ") for GeographyAreapageRender returned ": err})
				setStatusCode(req, w, err)
				return
			}
			var datasetlist models.DatasetList
			err = json.Unmarshal(b, &datasetlist)
			if err != nil {
				log.Error(err, log.Data{"error Unmarshaling the .Body data from the code-lists api http.Get(" + AreaPageLink + ") for GeographyAreapageRender returned ": err})
				setStatusCode(req, w, err)
				return
			}

			geographyDatasetLabel := ""
			geographyDatasetID := ""
			AreaPageMetadataLink := ""

			var geographyTypes []geographyAreaPage.AreaType
			for i := range datasetlistresults.Items {

				AreaPageMetadataLink = datasetlistresults.Items[i].Editions[0].Links.Latest.Href + `/metadata`

				resp, err := http.Get(AreaPageMetadataLink)
				if err != nil {
					log.Error(err, log.Data{"error getting metadata from the code-lists api http.Get(" + AreaPageMetadataLink + ") for GeographyAreapageRender returned ": err})
					setStatusCode(req, w, err)
					return
				}
				b, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Error(err, log.Data{"error getting the .Body data from the code-lists api http.Get(" + AreaPageMetadataLink + ") for GeographyAreapageRender returned ": err})
					setStatusCode(req, w, err)
					return
				}
				var datasetmetadataresults models.DatasetMetadata
				err = json.Unmarshal(b, &datasetmetadataresults)
				if err != nil {
					log.Error(err, log.Data{"error Unmarshaling the .Body data from the code-lists api http.Get(" + AreaPageMetadataLink + ") for GeographyAreapageRender returned ": err})
					setStatusCode(req, w, err)
					return
				}

				geographyDatasetLabel = datasetmetadataresults.Title
				geographyDatasetID = datasetmetadataresults.Description
				geographyTypes = append(geographyTypes, geographyAreaPage.AreaType{
					Label: geographyDatasetLabel,
					ID:    geographyDatasetID,
				})
			}

			page.Data.AreaTypes = geographyTypes
		} //err == nil
		page.Metadata.Title = areaTypeID
		page.DatasetTitle = datasetLabel
		page.DatasetId = datasetID

		templateJSON, err := json.Marshal(page)
		if err != nil {
			log.Error(err, log.Data{"error Marshaling the page data for GeographyAreapageRender returned ": err})
			setStatusCode(req, w, err)
			return
		}
		templateHTML, err := rend.Do("geography-area-page", templateJSON)
		if err != nil {
			log.Error(err, log.Data{"error rendering the geography-area-page data from GeographyAreapageRender returned ": err})
			setStatusCode(req, w, err)
			return
		}

		w.Write(templateHTML)
		return
	}
}
