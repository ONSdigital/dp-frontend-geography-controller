package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-frontend-models/model/geography/area"

	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/geography/list"
	"github.com/ONSdigital/go-ns/clients/codelist"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

type testCliError struct{}

func (e *testCliError) Error() string { return "client error" }
func (e *testCliError) Code() int     { return http.StatusNotFound }

func TestHandler(t *testing.T) {

	Convey("test setStatusCode", t, func() {

		Convey("test status code handles 404 response from client", func() {
			req := httptest.NewRequest("GET", "/geography/local-authority", nil)
			w := httptest.NewRecorder()
			err := &testCliError{}

			setStatusCode(req, w, err)

			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("test status code handles internal server error", func() {
			req := httptest.NewRequest("GET", "/geography/local-authority", nil)
			w := httptest.NewRecorder()
			err := errors.New("internal server error")

			setStatusCode(req, w, err)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

	Convey("test list page handler", t, func() {
		req, _ := http.NewRequest("GET", "/geography/local-authority", nil)
		w := httptest.NewRecorder()
		router := mux.NewRouter()

		Convey("sends data to the correct renderer endpoint", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(codeListID string) (codelist.EditionsListResults, error) {
					return codelist.EditionsListResults{}, nil
				},
				GetCodesFunc: func(codeListID string, edition string) (codelist.CodesResults, error) {
					return codelist.CodesResults{}, nil
				},
			}

			router.Path("/geography/{codeListID}").HandlerFunc(ListPageRender(mockRenderClient, mockCodeListClient))
			router.ServeHTTP(w, req)
			renderCall := mockRenderClient.DoCalls()[0]
			So(renderCall.In1, ShouldEqual, "geography-list")
		})

		Convey("maps the responses from the code list API to the frontend models", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(codeListID string) (codelist.EditionsListResults, error) {
					return codelist.EditionsListResults{
						Items: []codelist.EditionsList{
							codelist.EditionsList{
								Edition: "2018",
								Label:   "Local authority districts",
								Links:   codelist.EditionsListLink{},
							},
						},
						Count:      1,
						Offset:     0,
						Limit:      1,
						TotalCount: 1,
					}, nil
				},
				GetCodesFunc: func(codeListID string, edition string) (codelist.CodesResults, error) {
					return codelist.CodesResults{
						Items: []codelist.Item{
							codelist.Item{
								ID:    "E06000028",
								Label: "Bournemouth",
								Links: codelist.CodeLinks{},
							},
							codelist.Item{
								ID:    "S12000033",
								Label: "Aberdeen City",
								Links: codelist.CodeLinks{},
							},
							codelist.Item{
								ID:    "S12000034",
								Label: "Aberdeenshire",
								Links: codelist.CodeLinks{},
							},
						},
						Count:      2,
						TotalCount: 2,
						Offset:     0,
						Limit:      2,
					}, nil
				},
			}

			router.Path("/geography/{codeListID}").HandlerFunc(ListPageRender(mockRenderClient, mockCodeListClient))
			router.ServeHTTP(w, req)
			renderCall := mockRenderClient.DoCalls()[0]

			var payload list.Page
			err := json.Unmarshal(renderCall.In2, &payload)
			if err != nil {
				t.Errorf("Failed to unmarshal payload sent to renderer to relevant frontend model: %s", err)
				return
			}

			So(payload.Metadata.Title, ShouldEqual, "Local authority districts")
			So(payload.Breadcrumb, ShouldResemble, []model.TaxonomyNode{
				model.TaxonomyNode{
					Title: "Home",
					URI:   "https://www.ons.gov.uk",
				},
				model.TaxonomyNode{
					Title: "Geography",
					URI:   "/geography",
				},
				model.TaxonomyNode{
					Title: "Local authority districts",
					URI:   "/geography/local-authority",
				},
			})
			So(len(payload.Data.Items), ShouldEqual, 3)
			So(payload.Data.Items, ShouldResemble, []list.Item{
				list.Item{
					ID:    "S12000033",
					Label: "Aberdeen City",
					URI:   "/geography/local-authority/S12000033",
				},
				list.Item{
					ID:    "S12000034",
					Label: "Aberdeenshire",
					URI:   "/geography/local-authority/S12000034",
				},
				list.Item{
					ID:    "E06000028",
					Label: "Bournemouth",
					URI:   "/geography/local-authority/E06000028",
				},
			})
		})

		Convey("return an error status if request to GET code-list's editions fails", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(codeListID string) (codelist.EditionsListResults, error) {
					return codelist.EditionsListResults{}, errors.New("Code-list %s not found")
				},
				GetCodesFunc: func(codeListID string, edition string) (codelist.CodesResults, error) {
					return codelist.CodesResults{}, errors.New("Code-list %s not found")
				},
			}

			router.Path("/geography/{codeListID}").HandlerFunc(ListPageRender(mockRenderClient, mockCodeListClient))
			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 500)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 0)
		})

		Convey("return a 500 status if request to GET code-list's codes fails", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(codeListID string) (codelist.EditionsListResults, error) {
					return codelist.EditionsListResults{
						Items: []codelist.EditionsList{
							codelist.EditionsList{
								Edition: "2018",
								Label:   "Local authority districts",
								Links:   codelist.EditionsListLink{},
							},
						},
						Count:      1,
						Offset:     0,
						Limit:      1,
						TotalCount: 1,
					}, nil
				},
				GetCodesFunc: func(codeListID string, edition string) (codelist.CodesResults, error) {
					return codelist.CodesResults{}, errors.New("Code-list %s not found")
				},
			}

			router.Path("/geography/{codeListID}").HandlerFunc(ListPageRender(mockRenderClient, mockCodeListClient))
			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 500)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 0)
		})

		Convey("returns an error status if rendering service isn't responding", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return nil, errors.New("Unrecognised payload format")
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(codeListID string) (codelist.EditionsListResults, error) {
					return codelist.EditionsListResults{}, nil
				},
				GetCodesFunc: func(codeListID string, edition string) (codelist.CodesResults, error) {
					return codelist.CodesResults{}, nil
				},
			}

			router.Path("/geography/{codeListID}").HandlerFunc(ListPageRender(mockRenderClient, mockCodeListClient))
			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 500)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 1)
		})
	})

	Convey("test area page handler", t, func() {
		req, _ := http.NewRequest("GET", "/geography/local-authority/E07000223", nil)
		w := httptest.NewRecorder()
		router := mux.NewRouter()

		Convey("sends data to the correct renderer endpoint", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(codeListID string) (codelist.EditionsListResults, error) {
					return codelist.EditionsListResults{}, nil
				},
				GetCodeByIDFunc: func(codeListID string, edition string, codeID string) (codelist.CodeResult, error) {
					return codelist.CodeResult{}, nil
				},
				GetDatasetsByCodeFunc: func(codeListID string, edition string, codeID string) (codelist.DatasetsResult, error) {
					return codelist.DatasetsResult{}, nil
				},
			}
			mockDatasetClient := &DatasetClientMock{
				GetFunc: func(ctx context.Context, id string) (dataset.Model, error) {
					return dataset.Model{}, nil
				},
			}

			router.Path("/geography/{codeListID}/{codeID}").HandlerFunc(AreaPageRender(mockRenderClient, mockCodeListClient, mockDatasetClient))
			router.ServeHTTP(w, req)
			renderCall := mockRenderClient.DoCalls()[0]
			So(renderCall.In1, ShouldEqual, "geography-area")
		})

		Convey("maps the responses from the code list API to the frontend models", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(codeListID string) (codelist.EditionsListResults, error) {
					return codelist.EditionsListResults{
						Items: []codelist.EditionsList{
							codelist.EditionsList{
								Edition: "2018",
								Label:   "Local authority districts",
								Links:   codelist.EditionsListLink{},
							},
						},
						Count:      1,
						Offset:     0,
						Limit:      1,
						TotalCount: 1,
					}, nil
				},
				GetCodeByIDFunc: func(codeListID string, edition string, codeID string) (codelist.CodeResult, error) {
					return codelist.CodeResult{
						ID:    "E07000223",
						Label: "Adur",
					}, nil
				},
				GetDatasetsByCodeFunc: func(codeListID string, edition string, codeID string) (codelist.DatasetsResult, error) {
					return codelist.DatasetsResult{
						Datasets: []codelist.Dataset{
							codelist.Dataset{
								Links:          codelist.DatasetLinks{},
								DimensionLabal: "Adur",
								Editions: []codelist.DatasetEdition{
									codelist.DatasetEdition{
										Links: codelist.DatasetEditionLink{
											Self:            codelist.Link{},
											DatasetDimenion: codelist.Link{},
											LatestVersion: codelist.Link{
												ID:   "1",
												Href: "http://localhost:22000/datasets/mid-year-pop-est/editions/time-series/versions/1",
											},
										},
									},
								},
							},
						},
						Count: 1,
					}, nil
				},
			}
			mockDatasetClient := &DatasetClientMock{
				GetFunc: func(ctx context.Context, id string) (dataset.Model, error) {
					return dataset.Model{
						Description: "Test dataset description",
						Title:       "Test dataset title",
					}, nil
				},
			}

			router.Path("/geography/{codeListID}/{codeID}").HandlerFunc(AreaPageRender(mockRenderClient, mockCodeListClient, mockDatasetClient))
			router.ServeHTTP(w, req)
			renderCall := mockRenderClient.DoCalls()[0]

			var payload area.Page
			err := json.Unmarshal(renderCall.In2, &payload)
			if err != nil {
				t.Errorf("Failed to unmarshal payload sent to renderer to relevant frontend model: %s", err)
				return
			}

			So(payload.Metadata.Title, ShouldEqual, "Adur")
			So(payload.Breadcrumb, ShouldResemble, []model.TaxonomyNode{
				model.TaxonomyNode{
					Title: "Home",
					URI:   "https://www.ons.gov.uk",
				},
				model.TaxonomyNode{
					Title: "Geography",
					URI:   "/geography",
				},
				model.TaxonomyNode{
					Title: "Local authority districts",
					URI:   "/geography/local-authority",
				},
				model.TaxonomyNode{
					Title: "Adur",
					URI:   "/geography/local-authority/E07000223",
				},
			})
			So(len(payload.Data.Datasets), ShouldEqual, 1)
			So(payload.Data.Datasets, ShouldResemble, []area.Dataset{
				area.Dataset{
					ID:          "",
					Label:       "Test dataset title",
					Description: "Test dataset description",
					URI:         "/datasets/mid-year-pop-est/editions/time-series/versions/1",
				},
			})
		})

		Convey("return a 500 status if request to GET code-list's editions fails", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(codeListID string) (codelist.EditionsListResults, error) {
					return codelist.EditionsListResults{}, errors.New("Code-list %s not found")
				},
				GetCodeByIDFunc: func(codeListID string, edition string, codeID string) (codelist.CodeResult, error) {
					return codelist.CodeResult{}, nil
				},
				GetDatasetsByCodeFunc: func(codeListID string, edition string, codeID string) (codelist.DatasetsResult, error) {
					return codelist.DatasetsResult{}, nil
				},
			}
			mockDatasetClient := &DatasetClientMock{
				GetFunc: func(ctx context.Context, id string) (dataset.Model, error) {
					return dataset.Model{}, nil
				},
			}

			router.Path("/geography/{codeListID}/{codeID}").HandlerFunc(AreaPageRender(mockRenderClient, mockCodeListClient, mockDatasetClient))
			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 500)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 0)
		})

		Convey("return a 500 status if request to GET code-list's codes fails", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(codeListID string) (codelist.EditionsListResults, error) {
					return codelist.EditionsListResults{
						Items: []codelist.EditionsList{
							codelist.EditionsList{
								Edition: "2018",
								Label:   "Local authority districts",
								Links:   codelist.EditionsListLink{},
							},
						},
						Count:      1,
						Offset:     0,
						Limit:      1,
						TotalCount: 1,
					}, nil
				},
				GetCodeByIDFunc: func(codeListID string, edition string, codeID string) (codelist.CodeResult, error) {
					return codelist.CodeResult{}, errors.New("Code-list %s not found")
				},
				GetDatasetsByCodeFunc: func(codeListID string, edition string, codeID string) (codelist.DatasetsResult, error) {
					return codelist.DatasetsResult{}, nil
				},
			}
			mockDatasetClient := &DatasetClientMock{
				GetFunc: func(ctx context.Context, id string) (dataset.Model, error) {
					return dataset.Model{}, nil
				},
			}

			router.Path("/geography/{codeListID}/{codeID}").HandlerFunc(AreaPageRender(mockRenderClient, mockCodeListClient, mockDatasetClient))
			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 500)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 0)
		})

		Convey("return a 500 status if request to GET datasets by code fails", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(codeListID string) (codelist.EditionsListResults, error) {
					return codelist.EditionsListResults{
						Items: []codelist.EditionsList{
							codelist.EditionsList{
								Edition: "2018",
								Label:   "Local authority districts",
								Links:   codelist.EditionsListLink{},
							},
						},
						Count:      1,
						Offset:     0,
						Limit:      1,
						TotalCount: 1,
					}, nil
				},
				GetCodeByIDFunc: func(codeListID string, edition string, codeID string) (codelist.CodeResult, error) {
					return codelist.CodeResult{}, nil
				},
				GetDatasetsByCodeFunc: func(codeListID string, edition string, codeID string) (codelist.DatasetsResult, error) {
					return codelist.DatasetsResult{}, errors.New("Code-list %s not found")
				},
			}
			mockDatasetClient := &DatasetClientMock{
				GetFunc: func(ctx context.Context, id string) (dataset.Model, error) {
					return dataset.Model{}, nil
				},
			}

			router.Path("/geography/{codeListID}/{codeID}").HandlerFunc(AreaPageRender(mockRenderClient, mockCodeListClient, mockDatasetClient))
			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 500)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 0)
		})

		Convey("return a 500 status if request to GET dataset fails", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(codeListID string) (codelist.EditionsListResults, error) {
					return codelist.EditionsListResults{
						Items: []codelist.EditionsList{
							codelist.EditionsList{
								Edition: "2018",
								Label:   "Local authority districts",
								Links:   codelist.EditionsListLink{},
							},
						},
						Count:      1,
						Offset:     0,
						Limit:      1,
						TotalCount: 1,
					}, nil
				},
				GetCodeByIDFunc: func(codeListID string, edition string, codeID string) (codelist.CodeResult, error) {
					return codelist.CodeResult{}, nil
				},
				GetDatasetsByCodeFunc: func(codeListID string, edition string, codeID string) (codelist.DatasetsResult, error) {
					return codelist.DatasetsResult{
						Datasets: []codelist.Dataset{
							codelist.Dataset{
								Links:          codelist.DatasetLinks{},
								DimensionLabal: "Adur",
								Editions: []codelist.DatasetEdition{
									codelist.DatasetEdition{
										Links: codelist.DatasetEditionLink{
											Self:            codelist.Link{},
											DatasetDimenion: codelist.Link{},
											LatestVersion: codelist.Link{
												ID:   "1",
												Href: "http://localhost:22000/datasets/mid-year-pop-est/editions/time-series/versions/1",
											},
										},
									},
								},
							},
						},
						Count: 1,
					}, nil
				},
			}
			mockDatasetClient := &DatasetClientMock{
				GetFunc: func(ctx context.Context, id string) (dataset.Model, error) {
					return dataset.Model{}, errors.New("Dataset %s not found")
				},
			}

			router.Path("/geography/{codeListID}/{codeID}").HandlerFunc(AreaPageRender(mockRenderClient, mockCodeListClient, mockDatasetClient))
			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 500)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 0)
		})

		Convey("return a 500 status if rendering service isn't responding", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return nil, errors.New("Unrecognised payload format")
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(codeListID string) (codelist.EditionsListResults, error) {
					return codelist.EditionsListResults{}, nil
				},
				GetCodeByIDFunc: func(codeListID string, edition string, codeID string) (codelist.CodeResult, error) {
					return codelist.CodeResult{}, nil
				},
				GetDatasetsByCodeFunc: func(codeListID string, edition string, codeID string) (codelist.DatasetsResult, error) {
					return codelist.DatasetsResult{}, nil
				},
			}
			mockDatasetClient := &DatasetClientMock{
				GetFunc: func(ctx context.Context, id string) (dataset.Model, error) {
					return dataset.Model{}, nil
				},
			}

			router.Path("/geography/{codeListID}/{codeID}").HandlerFunc(AreaPageRender(mockRenderClient, mockCodeListClient, mockDatasetClient))
			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 500)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 1)
		})
	})
}
