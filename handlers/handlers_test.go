package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/codelist"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/headers"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/geography/area"
	"github.com/ONSdigital/dp-frontend-models/model/geography/list"
	"github.com/ONSdigital/go-ns/common"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	userAccessToken    = "Blackened is the end Winter it will send"
	serviceAccessToken = "Death of mother earth Never a rebirth Evolution's end Never will it mend never"
)

type testCliError struct{}

func (e *testCliError) Error() string { return "client error" }
func (e *testCliError) Code() int     { return http.StatusNotFound }

func assertAuthTokens(actualUserToken, actualServiceToken string) {
	So(actualUserToken, ShouldEqual, userAccessToken)
	So(actualServiceToken, ShouldEqual, serviceAccessToken)
}

func TestHandler(t *testing.T) {

	Convey("test setStatusCode", t, func() {

		Convey("test status code handles 404 response from client", func() {
			req := httptest.NewRequest("GET", "/foobar", nil)
			w := httptest.NewRecorder()
			err := &testCliError{}

			setStatusCode(req, w, err)

			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("test status code handles internal server error", func() {
			req := httptest.NewRequest("GET", "/foobar", nil)
			w := httptest.NewRecorder()
			err := errors.New("internal server error")

			setStatusCode(req, w, err)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})

	Convey("test Homepage handler", t, func() {
		req, _ := http.NewRequest("GET", "/geography", nil)
		headers.SetUserAuthToken(req, userAccessToken)
		headers.SetServiceAuthToken(req, serviceAccessToken)

		w := httptest.NewRecorder()
		router := mux.NewRouter()

		Convey("sends data to the correct renderer endpoint", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetGeographyCodeListsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string) (codelist.CodeListResults, error) {
					return codelist.CodeListResults{}, nil
				},
			}

			router.Path("/geography").HandlerFunc(HomepageRender(mockRenderClient, mockCodeListClient))

			router.ServeHTTP(w, req)

			renderCall := mockRenderClient.DoCalls()[0]
			So(renderCall.In1, ShouldEqual, "geography-homepage")

			calls := mockCodeListClient.GetGeographyCodeListsCalls()
			So(calls, ShouldHaveLength, 1)
			assertAuthTokens(calls[0].UserAuthToken, calls[0].ServiceAuthToken)
		})

		Convey("return a 404 status if request to GET code-list return's a 404", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetGeographyCodeListsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string) (codelist.CodeListResults, error) {
					return codelist.CodeListResults{}, &testCliError{}
				},
			}

			router.Path("/geography").HandlerFunc(HomepageRender(mockRenderClient, mockCodeListClient))

			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, 404)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 0)

			calls := mockCodeListClient.GetGeographyCodeListsCalls()
			So(calls, ShouldHaveLength, 1)
			assertAuthTokens(calls[0].UserAuthToken, calls[0].ServiceAuthToken)
		})

		Convey("return a 500 status if request to GET code-list's codes fails", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetGeographyCodeListsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string) (codelist.CodeListResults, error) {
					return codelist.CodeListResults{}, errors.New("Code-list %s not found")
				},
			}

			router.Path("/geography").HandlerFunc(HomepageRender(mockRenderClient, mockCodeListClient))
			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, 500)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 0)

			calls := mockCodeListClient.GetGeographyCodeListsCalls()
			So(calls, ShouldHaveLength, 1)
			assertAuthTokens(calls[0].UserAuthToken, calls[0].ServiceAuthToken)
		})

		Convey("return a 500 status if rendering service doesn't respond", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return nil, errors.New("Unrecognised payload format")
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetGeographyCodeListsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string) (codelist.CodeListResults, error) {
					return codelist.CodeListResults{}, nil
				},
			}

			router.Path("/geography").HandlerFunc(HomepageRender(mockRenderClient, mockCodeListClient))
			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 500)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 1)

			calls := mockCodeListClient.GetGeographyCodeListsCalls()
			So(calls, ShouldHaveLength, 1)
			assertAuthTokens(calls[0].UserAuthToken, calls[0].ServiceAuthToken)
		})
	})

	Convey("test list page handler", t, func() {
		req, _ := http.NewRequest("GET", "/geography/local-authority", nil)
		headers.SetUserAuthToken(req, userAccessToken)
		headers.SetServiceAuthToken(req, serviceAccessToken)
		w := httptest.NewRecorder()
		router := mux.NewRouter()

		Convey("sends data to the correct renderer endpoint", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string) (codelist.EditionsListResults, error) {
					return codelist.EditionsListResults{}, nil
				},
				GetCodesFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string) (codelist.CodesResults, error) {
					return codelist.CodesResults{}, nil
				},
			}

			router.Path("/geography/{codeListID}").HandlerFunc(ListPageRender(mockRenderClient, mockCodeListClient))

			router.ServeHTTP(w, req)

			renderCall := mockRenderClient.DoCalls()[0]
			So(renderCall.In1, ShouldEqual, "geography-list")

			calls := mockCodeListClient.GetCodeListEditionsCalls()
			So(calls, ShouldHaveLength, 1)
			assertAuthTokens(calls[0].UserAuthToken, calls[0].ServiceAuthToken)

			getCodesCalls := mockCodeListClient.GetCodesCalls()
			So(getCodesCalls, ShouldHaveLength, 0)
		})

		Convey("maps the responses from the code list API to the frontend models", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string) (codelist.EditionsListResults, error) {
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
				GetCodesFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string) (codelist.CodesResults, error) {
					return codelist.CodesResults{
						Items: []codelist.Item{
							codelist.Item{
								Code:  "E06000028",
								Label: "Bournemouth",
								Links: codelist.CodeLinks{},
							},
							codelist.Item{
								Code:  "S12000033",
								Label: "Aberdeen City",
								Links: codelist.CodeLinks{},
							},
							codelist.Item{
								Code:  "S12000034",
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
			if err := json.Unmarshal(renderCall.In2, &payload); err != nil {
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

			calls := mockCodeListClient.GetCodeListEditionsCalls()
			So(calls, ShouldHaveLength, 1)
			assertAuthTokens(calls[0].UserAuthToken, calls[0].ServiceAuthToken)

			getCodesCalls := mockCodeListClient.GetCodesCalls()
			So(getCodesCalls, ShouldHaveLength, 1)
			assertAuthTokens(getCodesCalls[0].UserAuthToken, getCodesCalls[0].ServiceAuthToken)
		})

		Convey("return an error status if request to GET code-list's editions fails", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string) (codelist.EditionsListResults, error) {
					return codelist.EditionsListResults{}, errors.New("Code-list %s not found")
				},
				GetCodesFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string) (codelist.CodesResults, error) {
					return codelist.CodesResults{}, errors.New("Code-list %s not found")
				},
			}

			router.Path("/geography/{codeListID}").HandlerFunc(ListPageRender(mockRenderClient, mockCodeListClient))
			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, 500)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 0)

			calls := mockCodeListClient.GetCodeListEditionsCalls()
			So(calls, ShouldHaveLength, 1)
			assertAuthTokens(calls[0].UserAuthToken, calls[0].ServiceAuthToken)

			getCodesCalls := mockCodeListClient.GetCodesCalls()
			So(getCodesCalls, ShouldBeNil)
		})

		Convey("return a 500 status if request to GET code-list's codes fails", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string) (codelist.EditionsListResults, error) {
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
				GetCodesFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string) (codelist.CodesResults, error) {
					return codelist.CodesResults{}, errors.New("Code-list %s not found")
				},
			}

			router.Path("/geography/{codeListID}").HandlerFunc(ListPageRender(mockRenderClient, mockCodeListClient))
			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, 500)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 0)

			calls := mockCodeListClient.GetCodeListEditionsCalls()
			So(calls, ShouldHaveLength, 1)
			assertAuthTokens(calls[0].UserAuthToken, calls[0].ServiceAuthToken)

			getCodesCalls := mockCodeListClient.GetCodesCalls()
			So(getCodesCalls, ShouldHaveLength, 1)
			assertAuthTokens(getCodesCalls[0].UserAuthToken, getCodesCalls[0].ServiceAuthToken)
		})

		Convey("return a 500 status if rendering service doesn't respond", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return nil, errors.New("Unrecognised payload format")
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string) (codelist.EditionsListResults, error) {
					return codelist.EditionsListResults{}, nil
				},
				GetCodesFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string) (codelist.CodesResults, error) {
					return codelist.CodesResults{}, nil
				},
			}

			router.Path("/geography/{codeListID}").HandlerFunc(ListPageRender(mockRenderClient, mockCodeListClient))
			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, 500)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 1)

			calls := mockCodeListClient.GetCodeListEditionsCalls()
			So(calls, ShouldHaveLength, 1)
			assertAuthTokens(calls[0].UserAuthToken, calls[0].ServiceAuthToken)

			getCodesCalls := mockCodeListClient.GetCodesCalls()
			So(getCodesCalls, ShouldBeNil)
		})
	})
}

func TestAreaPageRender(t *testing.T) {
	Convey("test area page handler", t, func() {
		req, _ := http.NewRequest("GET", "/geography/local-authority/E07000223", nil)
		headers.SetUserAuthToken(req, userAccessToken)
		headers.SetServiceAuthToken(req, serviceAccessToken)
		w := httptest.NewRecorder()
		router := mux.NewRouter()

		Convey("sends data to the correct renderer endpoint", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string) (codelist.EditionsListResults, error) {
					return codelist.EditionsListResults{}, nil
				},
				GetCodeByIDFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (codelist.CodeResult, error) {
					return codelist.CodeResult{}, nil
				},
				GetDatasetsByCodeFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (codelist.DatasetsResult, error) {
					return codelist.DatasetsResult{}, nil
				},
			}
			mockDatasetClient := &DatasetClientMock{
				GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (i dataset.DatasetDetails, e error) {
					return dataset.DatasetDetails{}, nil
				},
			}

			router.Path("/geography/{codeListID}/{codeID}").HandlerFunc(AreaPageRender(mockRenderClient, mockCodeListClient, mockDatasetClient))
			router.ServeHTTP(w, req)

			renderCall := mockRenderClient.DoCalls()[0]
			So(renderCall.In1, ShouldEqual, "geography-area")

			Convey("and the expected requests are made to the codelist API are made", func() {
				editionCalls := mockCodeListClient.GetCodeListEditionsCalls()
				So(editionCalls, ShouldHaveLength, 1)
				assertAuthTokens(editionCalls[0].UserAuthToken, editionCalls[0].ServiceAuthToken)

				So(mockCodeListClient.GetCodeByIDCalls(), ShouldBeNil)
				So(mockCodeListClient.GetDatasetsByCodeCalls(), ShouldBeNil)
			})

		})

		Convey("maps the responses from the code list API to the frontend models", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string) (codelist.EditionsListResults, error) {
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
				GetCodeByIDFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (codelist.CodeResult, error) {
					return codelist.CodeResult{
						ID:    "E07000223",
						Label: "Adur",
					}, nil
				},
				GetDatasetsByCodeFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (codelist.DatasetsResult, error) {
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
				GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (i dataset.DatasetDetails, e error) {
					return dataset.DatasetDetails{
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

			Convey("the expected requests are made to the codelist API", func() {
				editionCalls := mockCodeListClient.GetCodeListEditionsCalls()
				So(editionCalls, ShouldHaveLength, 1)
				assertAuthTokens(editionCalls[0].UserAuthToken, editionCalls[0].ServiceAuthToken)

				getCodeByIdCalls := mockCodeListClient.GetCodeByIDCalls()
				So(getCodeByIdCalls, ShouldHaveLength, 1)
				assertAuthTokens(getCodeByIdCalls[0].UserAuthToken, getCodeByIdCalls[0].ServiceAuthToken)

				datasetsByCodeCalls := mockCodeListClient.GetDatasetsByCodeCalls()
				So(datasetsByCodeCalls, ShouldHaveLength, 1)
				assertAuthTokens(datasetsByCodeCalls[0].UserAuthToken, datasetsByCodeCalls[0].ServiceAuthToken)
			})
		})

		Convey("return a 500 status if request to GET code-list's editions fails", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string) (codelist.EditionsListResults, error) {
					return codelist.EditionsListResults{}, errors.New("Code-list %s not found")
				},
				GetCodeByIDFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (codelist.CodeResult, error) {
					return codelist.CodeResult{}, nil
				},
				GetDatasetsByCodeFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (codelist.DatasetsResult, error) {
					return codelist.DatasetsResult{}, nil
				},
			}
			mockDatasetClient := &DatasetClientMock{
				GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (i dataset.DatasetDetails, e error) {
					return dataset.DatasetDetails{}, nil
				},
			}

			router.Path("/geography/{codeListID}/{codeID}").HandlerFunc(AreaPageRender(mockRenderClient, mockCodeListClient, mockDatasetClient))
			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 500)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 0)

			Convey("the expected requests are made to the codelist API", func() {
				editionCalls := mockCodeListClient.GetCodeListEditionsCalls()
				So(editionCalls, ShouldHaveLength, 1)
				assertAuthTokens(editionCalls[0].UserAuthToken, editionCalls[0].ServiceAuthToken)

				getCodeByIdCalls := mockCodeListClient.GetCodeByIDCalls()
				So(getCodeByIdCalls, ShouldBeNil)

				datasetsByCodeCalls := mockCodeListClient.GetDatasetsByCodeCalls()
				So(datasetsByCodeCalls, ShouldBeNil)
			})
		})

		Convey("return a 500 status if request to GET code-list's codes fails", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string) (codelist.EditionsListResults, error) {
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
				GetCodeByIDFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (codelist.CodeResult, error) {
					return codelist.CodeResult{}, errors.New("Code-list %s not found")
				},
				GetDatasetsByCodeFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (codelist.DatasetsResult, error) {
					return codelist.DatasetsResult{}, nil
				},
			}
			mockDatasetClient := &DatasetClientMock{
				GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (i dataset.DatasetDetails, e error) {
					return dataset.DatasetDetails{}, nil
				},
			}

			router.Path("/geography/{codeListID}/{codeID}").HandlerFunc(AreaPageRender(mockRenderClient, mockCodeListClient, mockDatasetClient))
			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 500)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 0)

			Convey("the expected requests are made to the codelist API", func() {
				editionCalls := mockCodeListClient.GetCodeListEditionsCalls()
				So(editionCalls, ShouldHaveLength, 1)
				assertAuthTokens(editionCalls[0].UserAuthToken, editionCalls[0].ServiceAuthToken)

				getCodeByIdCalls := mockCodeListClient.GetCodeByIDCalls()
				So(getCodeByIdCalls, ShouldHaveLength, 1)
				assertAuthTokens(getCodeByIdCalls[0].UserAuthToken, getCodeByIdCalls[0].ServiceAuthToken)

				datasetsByCodeCalls := mockCodeListClient.GetDatasetsByCodeCalls()
				So(datasetsByCodeCalls, ShouldHaveLength, 0)
			})
		})

		Convey("return a 500 status if request to GET datasets by code fails", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string) (codelist.EditionsListResults, error) {
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
				GetCodeByIDFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (codelist.CodeResult, error) {
					return codelist.CodeResult{}, nil
				},
				GetDatasetsByCodeFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (codelist.DatasetsResult, error) {
					return codelist.DatasetsResult{}, errors.New("Code-list %s not found")
				},
			}
			mockDatasetClient := &DatasetClientMock{
				GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (i dataset.DatasetDetails, e error) {
					return dataset.DatasetDetails{}, nil
				},
			}

			router.Path("/geography/{codeListID}/{codeID}").HandlerFunc(AreaPageRender(mockRenderClient, mockCodeListClient, mockDatasetClient))
			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 500)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 0)

			Convey("the expected requests are made to the codelist API", func() {
				editionCalls := mockCodeListClient.GetCodeListEditionsCalls()
				So(editionCalls, ShouldHaveLength, 1)
				assertAuthTokens(editionCalls[0].UserAuthToken, editionCalls[0].ServiceAuthToken)

				getCodeByIdCalls := mockCodeListClient.GetCodeByIDCalls()
				So(getCodeByIdCalls, ShouldHaveLength, 1)
				assertAuthTokens(getCodeByIdCalls[0].UserAuthToken, getCodeByIdCalls[0].ServiceAuthToken)

				datasetsByCodeCalls := mockCodeListClient.GetDatasetsByCodeCalls()
				So(datasetsByCodeCalls, ShouldHaveLength, 1)
				assertAuthTokens(datasetsByCodeCalls[0].UserAuthToken, datasetsByCodeCalls[0].ServiceAuthToken)
			})
		})

		Convey("return a 500 status if request to GET dataset fails", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return bytes, nil
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string) (codelist.EditionsListResults, error) {
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
				GetCodeByIDFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (codelist.CodeResult, error) {
					return codelist.CodeResult{}, nil
				},
				GetDatasetsByCodeFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (codelist.DatasetsResult, error) {
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
				GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (i dataset.DatasetDetails, e error) {
					return dataset.DatasetDetails{}, errors.New("Dataset %s not found")
				},
			}

			router.Path("/geography/{codeListID}/{codeID}").HandlerFunc(AreaPageRender(mockRenderClient, mockCodeListClient, mockDatasetClient))
			router.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, 500)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 0)

			Convey("the expected requests are made to the codelist API", func() {
				editionCalls := mockCodeListClient.GetCodeListEditionsCalls()
				So(editionCalls, ShouldHaveLength, 1)
				assertAuthTokens(editionCalls[0].UserAuthToken, editionCalls[0].ServiceAuthToken)

				getCodeByIdCalls := mockCodeListClient.GetCodeByIDCalls()
				So(getCodeByIdCalls, ShouldHaveLength, 1)
				assertAuthTokens(getCodeByIdCalls[0].UserAuthToken, getCodeByIdCalls[0].ServiceAuthToken)

				datasetsByCodeCalls := mockCodeListClient.GetDatasetsByCodeCalls()
				So(datasetsByCodeCalls, ShouldHaveLength, 1)
				assertAuthTokens(datasetsByCodeCalls[0].UserAuthToken, datasetsByCodeCalls[0].ServiceAuthToken)
			})
		})

		Convey("return a 500 status if rendering service isn't responding", func() {
			mockRenderClient := &RenderClientMock{
				DoFunc: func(path string, bytes []byte) ([]byte, error) {
					return nil, errors.New("Unrecognised payload format")
				},
			}
			mockCodeListClient := &CodeListClientMock{
				GetCodeListEditionsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string) (codelist.EditionsListResults, error) {
					return codelist.EditionsListResults{}, nil
				},
				GetCodeByIDFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (codelist.CodeResult, error) {
					return codelist.CodeResult{}, nil
				},
				GetDatasetsByCodeFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, odeListID string, edition string, codeID string) (codelist.DatasetsResult, error) {
					return codelist.DatasetsResult{}, nil
				},
			}
			mockDatasetClient := &DatasetClientMock{
				GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (i dataset.DatasetDetails, e error) {
					return dataset.DatasetDetails{}, nil
				},
			}

			router.Path("/geography/{codeListID}/{codeID}").HandlerFunc(AreaPageRender(mockRenderClient, mockCodeListClient, mockDatasetClient))
			router.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 500)
			So(len(mockRenderClient.DoCalls()), ShouldEqual, 1)

			Convey("the expected requests are made to the codelist API", func() {
				editionCalls := mockCodeListClient.GetCodeListEditionsCalls()
				So(editionCalls, ShouldHaveLength, 1)
				assertAuthTokens(editionCalls[0].UserAuthToken, editionCalls[0].ServiceAuthToken)

				getCodeByIdCalls := mockCodeListClient.GetCodeByIDCalls()
				So(getCodeByIdCalls, ShouldHaveLength, 0)

				datasetsByCodeCalls := mockCodeListClient.GetDatasetsByCodeCalls()
				So(datasetsByCodeCalls, ShouldHaveLength, 0)
			})
		})
	})
}

func TestGetUserAuthToken(t *testing.T) {
	Convey("should return X-Florence-Token value", t, func() {
		r, err := http.NewRequest(http.MethodGet, "http://localhost:8080/test", nil)
		headers.SetUserAuthToken(r, userAccessToken)
		So(err, ShouldBeNil)

		actual := getUserAuthToken(nil, r)
		So(actual, ShouldEqual, userAccessToken)
	})

	Convey("should return access_token cookie value", t, func() {
		r, err := http.NewRequest(http.MethodGet, "http://localhost:8080/test", nil)
		r.AddCookie(&http.Cookie{Name: common.FlorenceCookieKey, Value: common.FlorenceCookieKey})
		So(err, ShouldBeNil)

		actual := getUserAuthToken(nil, r)
		So(actual, ShouldEqual, common.FlorenceCookieKey)
	})

	Convey("should return empty if not set", t, func() {
		r, err := http.NewRequest(http.MethodGet, "http://localhost:8080/test", nil)
		So(err, ShouldBeNil)

		actual := getUserAuthToken(nil, r)
		So(actual, ShouldBeEmpty)
	})
}

func TestUnitMapCookiesPreferences(t *testing.T) {
	req := httptest.NewRequest("", "/", nil)
	pageModel := model.Page{
		CookiesPreferencesSet: false,
		CookiesPolicy: model.CookiesPolicy{
			Essential: false,
			Usage:     false,
		},
	}

	Convey("maps cookies preferences cookie data to page model correctly", t, func() {
		So(pageModel.CookiesPreferencesSet, ShouldEqual, false)
		So(pageModel.CookiesPolicy.Essential, ShouldEqual, false)
		So(pageModel.CookiesPolicy.Usage, ShouldEqual, false)
		req.AddCookie(&http.Cookie{Name: "cookies_preferences_set", Value: "true"})
		req.AddCookie(&http.Cookie{Name: "cookies_policy", Value: "%7B%22essential%22%3Atrue%2C%22usage%22%3Atrue%7D"})
		mapCookiePreferences(req, &pageModel.CookiesPreferencesSet, &pageModel.CookiesPolicy)
		So(pageModel.CookiesPreferencesSet, ShouldEqual, true)
		So(pageModel.CookiesPolicy.Essential, ShouldEqual, true)
		So(pageModel.CookiesPolicy.Usage, ShouldEqual, true)
	})
}
