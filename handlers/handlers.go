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

// const typeGeography = `/local-authority/editions`

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

//GeographyHomepageRender ...
func GeographyHomepageRender(rend RenderClient) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		var page geographyHomepage.Page

		resp, err := http.Get(`https://api.dev.cmd.onsdigital.co.uk/v1/code-lists` + typeGeography)
		if err != nil {
			log.Error(err, log.Data{"test": "error http.Get"})
			setStatusCode(req, w, err)
			return
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err, log.Data{"test": "error Body"})
			setStatusCode(req, w, err)
			return
		}
		var codelistresults models.CodeListResults
		err = json.Unmarshal(b, &codelistresults)
		if err != nil {
			log.Error(err, log.Data{"test": "error codelistresults"})
			setStatusCode(req, w, err)
			return
		}
		var codelist models.CodeList
		err = json.Unmarshal(b, &codelist)
		if err != nil {
			log.Error(err, log.Data{"test": "error codelist"})
			setStatusCode(req, w, err)
			return
		}

		// geographyTypesLabel := ""
		geographyTypesID := ""
		for i := range codelistresults.Items {
			log.Debug("test", log.Data{
				"geographyTypesTest": codelistresults.Items[i],
			})

			geographyTypesID = geographyTypesID + codelistresults.Items[i].Links.Self.ID
			// geographyTypesLabel = geographyTypesLabel + codelistresults.Items[i].Links.Editions.Href
			// resp2, err := http.Get(`https://api.dev.cmd.onsdigital.co.uk/v1/code-lists` + typeGeography)
			// if err != nil {
			// 	setStatusCode(req, w, err)
			// 	return
			// }
			// b2, err := ioutil.ReadAll(resp2.Body)
			// if err != nil {
			// 	setStatusCode(req, w, err)
			// 	return
			// }
			// var codelistresults2 models.CodeListResults
			// err = json.Unmarshal(b2, &codelistresults)
			// if err != nil {
			// 	setStatusCode(req, w, err)
			// 	return
			// }
			// geographyTypesLabel = codelistresults2

		}

		page.Data.AreaTypes = []geographyHomepage.AreaType{
			// {Label: "Countries", ID: "country"},
			// {Label: "Regions", ID: "region"},
			// {Label: "Local authorities", ID: "local-authority"},
			{Label: geographyTypesID, ID: geographyTypesID},
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

//GeographyListpageRender ...
func GeographyListpageRender(rend RenderClient) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		var page geographyListPage.Page

		vars := mux.Vars(req)
		areaTypeID := vars["areaTypeID"]

		resp, err := http.Get(`https://api.dev.cmd.onsdigital.co.uk/v1/code-lists/` + areaTypeID + `/editions/2016/codes`)
		// fmt.Println("error err", err)
		// fmt.Println("error resp.StatusCode", resp.StatusCode)
		if err != nil {
			log.Error(err, log.Data{"test": "error http.Get"})
			setStatusCode(req, w, err)
			return
		}
		if resp.StatusCode == 404 {
			// fmt.Println("ERROR 404 not 500")
			log.Error(err, log.Data{"test": "error resp"})
			setStatusCode(req, w, err) //<--error 500
			return
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err, log.Data{"test": "error Body"})
			setStatusCode(req, w, err)
			return
		}

		var codelistresults models.CodeListResults
		err = json.Unmarshal(b, &codelistresults)
		if err != nil {
			log.Error(err, log.Data{"test": "error codelistresults"})
			setStatusCode(req, w, err)
			return
		}
		var codelist models.CodeList
		err = json.Unmarshal(b, &codelist)
		if err != nil {
			log.Error(err, log.Data{"test": "error codelist"})
			setStatusCode(req, w, err)
			return
		}

		var geographyTypes []geographyListPage.AreaType
		for i := range codelistresults.Items {

			if i >= 10 {
				break
			}

			geographyTypes = append(geographyTypes, geographyListPage.AreaType{Label: codelistresults.Items[i].Label, ID: codelistresults.Items[i].ID})
		}

		page.Data.AreaTypes = geographyTypes
		page.Metadata.Title = areaTypeID

		templateJSON, err := json.Marshal(page)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}
		templateHTML, err := rend.Do("geography-list-page", templateJSON)
		if err != nil {
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

		resp, err := http.Get(`https://api.dev.cmd.onsdigital.co.uk/v1/code-lists/` + areaTypeID + `/editions/2016/codes/` + datasetID + `/datasets`)
		if err != nil {
			log.Error(err, log.Data{"test": "error http.Get"})
			// log.Debug("error resp.StatusCode ", resp.StatusCode)
			setStatusCode(req, w, err)
			return
		}
		if resp.StatusCode == 404 {
			log.Error(err, log.Data{"test": "error resp"})
			// fmt.Println("ERROR 404 not 500")
			// setStatusCode(req, w, err) //<--error 500
			// return
		} else {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error(err, log.Data{"test": "error Body"})
				// log.Debug("error Body ", err)
				setStatusCode(req, w, err)
				return
			}

			var codelistresults models.CodeListResults
			err = json.Unmarshal(b, &codelistresults)
			if err != nil {
				log.Error(err, log.Data{"test": "error codelistresults"})
				// log.Debug("error Unmarshal codelistresults ", err)
				setStatusCode(req, w, err)
				return
			}
			var codelist models.CodeList
			err = json.Unmarshal(b, &codelist)
			if err != nil {
				log.Error(err, log.Data{"test": "error CodeList"})
				// log.Debug("error Unmarshal codelist ", err)
				setStatusCode(req, w, err)
				return
			}

			var geographyTypes []geographyAreaPage.AreaType
			for i := range codelistresults.Items {

				if i >= 10 {
					break
				}

				// geographyTypes = append(geographyTypes, geographyAreaPage.AreaType{Label: codelistresults.Items[i].Links.Self.ID, ID: codelistresults.Items[i].editions[0].Links.Self.href})
				geographyTypes = append(geographyTypes, geographyAreaPage.AreaType{Label: codelistresults.Items[i].Links.Self.ID, ID: codelistresults.Items[i].Links.Self.Href})
			}

			page.Data.AreaTypes = geographyTypes
		} //if resp.StatusCode == 404
		page.Metadata.Title = areaTypeID
		page.DatasetTitle = datasetLabel
		page.DatasetId = datasetID

		templateJSON, err := json.Marshal(page)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}
		templateHTML, err := rend.Do("geography-area-page", templateJSON)
		if err != nil {
			setStatusCode(req, w, err)
			return
		}

		w.Write(templateHTML)
		return
	}
}
