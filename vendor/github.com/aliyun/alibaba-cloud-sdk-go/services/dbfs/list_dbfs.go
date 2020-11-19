package dbfs

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//
// Code generated by Alibaba Cloud SDK Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

// ListDbfs invokes the dbfs.ListDbfs API synchronously
func (client *Client) ListDbfs(request *ListDbfsRequest) (response *ListDbfsResponse, err error) {
	response = CreateListDbfsResponse()
	err = client.DoAction(request, response)
	return
}

// ListDbfsWithChan invokes the dbfs.ListDbfs API asynchronously
func (client *Client) ListDbfsWithChan(request *ListDbfsRequest) (<-chan *ListDbfsResponse, <-chan error) {
	responseChan := make(chan *ListDbfsResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.ListDbfs(request)
		if err != nil {
			errChan <- err
		} else {
			responseChan <- response
		}
	})
	if err != nil {
		errChan <- err
		close(responseChan)
		close(errChan)
	}
	return responseChan, errChan
}

// ListDbfsWithCallback invokes the dbfs.ListDbfs API asynchronously
func (client *Client) ListDbfsWithCallback(request *ListDbfsRequest, callback func(response *ListDbfsResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *ListDbfsResponse
		var err error
		defer close(result)
		response, err = client.ListDbfs(request)
		callback(response, err)
		result <- 1
	})
	if err != nil {
		defer close(result)
		callback(nil, err)
		result <- 0
	}
	return result
}

// ListDbfsRequest is the request struct for api ListDbfs
type ListDbfsRequest struct {
	*requests.RpcRequest
	SortType    string           `position:"Query" name:"SortType"`
	ClientToken string           `position:"Query" name:"ClientToken"`
	FilterValue string           `position:"Query" name:"FilterValue"`
	PageNumber  requests.Integer `position:"Query" name:"PageNumber"`
	Tags        string           `position:"Query" name:"Tags"`
	FilterKey   string           `position:"Query" name:"FilterKey"`
	SortKey     string           `position:"Query" name:"SortKey"`
	PageSize    requests.Integer `position:"Query" name:"PageSize"`
}

// ListDbfsResponse is the response struct for api ListDbfs
type ListDbfsResponse struct {
	*responses.BaseResponse
	RequestId  string `json:"RequestId" xml:"RequestId"`
	TotalCount int    `json:"TotalCount" xml:"TotalCount"`
	PageNumber int    `json:"PageNumber" xml:"PageNumber"`
	PageSize   int    `json:"PageSize" xml:"PageSize"`
	DBFSInfo   []Info `json:"DBFSInfo" xml:"DBFSInfo"`
}

// CreateListDbfsRequest creates a request to invoke ListDbfs API
func CreateListDbfsRequest() (request *ListDbfsRequest) {
	request = &ListDbfsRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("DBFS", "2020-04-18", "ListDbfs", "dbfs", "openAPI")
	request.Method = requests.POST
	return
}

// CreateListDbfsResponse creates a response to parse from ListDbfs response
func CreateListDbfsResponse() (response *ListDbfsResponse) {
	response = &ListDbfsResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}