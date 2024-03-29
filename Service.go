package facebook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	errortools "github.com/leapforce-libraries/go_errortools"
	go_http "github.com/leapforce-libraries/go_http"
	ig "github.com/leapforce-libraries/go_integration"
)

const (
	apiName               string = "facebook"
	apiUrl                string = "https://graph.facebook.com/v13.0"
	errorCodeTooManyCalls int    = 80004
)

// Service stores Service configuration
//
type Service struct {
	accessToken string
	httpService *go_http.Service
}

type ServiceConfig struct {
	AccessToken string
}

// methods
//
func NewService(serviceConfig *ServiceConfig) (*Service, *errortools.Error) {
	if serviceConfig == nil {
		return nil, errortools.ErrorMessage("ServiceConfig must not be a nil pointer")
	}

	if serviceConfig.AccessToken == "" {
		return nil, errortools.ErrorMessage("AccessToken not provided")
	}

	httpService, e := go_http.NewService(&go_http.ServiceConfig{})
	if e != nil {
		return nil, e
	}

	return &Service{
		accessToken: serviceConfig.AccessToken,
		httpService: httpService,
	}, nil
}

func (service *Service) httpRequest(requestConfig *go_http.RequestConfig) (*http.Request, *http.Response, *errortools.Error) {
	// add authentication header
	header := http.Header{}
	header.Set("Authorization", fmt.Sprintf("Bearer %s", service.accessToken))
	(*requestConfig).NonDefaultHeaders = &header

	// add error model
	errorResponse := ErrorResponse{}
	(*requestConfig).ErrorModel = &errorResponse

	request, response, e := service.httpService.HttpRequest(requestConfig)
	if errorResponse.Error.ErrorUserMsg != "" {
		if errorResponse.Error.ErrorUserTitle != "" {
			e.SetMessagef("%s: %s", errorResponse.Error.ErrorUserTitle, errorResponse.Error.ErrorUserMsg)
		} else {
			e.SetMessage(errorResponse.Error.ErrorUserMsg)
		}
	} else if errorResponse.Error.Message != "" {
		e.SetMessage(errorResponse.Error.Message)
	}

	if errorResponse.Error.Code == errorCodeTooManyCalls {
		fmt.Println("Waiting another minute...")
		time.Sleep(time.Minute)

		// retry
		return service.httpRequest(requestConfig)
	}

	if response != nil {
		header := response.Header.Get("x-business-use-case-usage")
		if header != "" {
			businessUseCaseUsage := xBusinessUseCaseUsage{}
			err := json.Unmarshal([]byte(header), &businessUseCaseUsage)
			if err != nil {
				errortools.CaptureError(err)
			} else {
				for _, b := range businessUseCaseUsage {
					if ig.Debug() {
						fmt.Printf("%+v\n", b)
					}
					if len(b) > 0 {
						estimatedTimeToRegainAccess := b[0].EstimatedTimeToRegainAccess
						if estimatedTimeToRegainAccess > 0 {
							fmt.Printf("Waiting %v minutes to regain access...\n", estimatedTimeToRegainAccess)
							time.Sleep(time.Minute * time.Duration(estimatedTimeToRegainAccess+1))

							// retry
							return service.httpRequest(requestConfig)
						}
					}
				}
			}
		}
	}

	return request, response, e
}

func (service *Service) url(path string) string {
	return fmt.Sprintf("%s/%s", apiUrl, path)
}

func (service *Service) AccessToken(accessToken string) {
	service.accessToken = accessToken
}

func (service Service) ApiName() string {
	return apiName
}

func (service Service) ApiKey() string {
	return service.accessToken
}

func (service Service) ApiCallCount() int64 {
	return service.httpService.RequestCount()
}

func (service Service) ApiReset() {
	service.httpService.ResetRequestCount()
}
