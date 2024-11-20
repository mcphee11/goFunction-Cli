
package start

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mypurecloud/platform-client-sdk-go/platformclientv2"
)

type RequestBody struct {
// build json struct
Id string `json:"Id"`
}

func genesysStart(w http.ResponseWriter, config *platformclientv2.Configuration, requestBody RequestBody) {
	// Set the API your wanting to use eg Conversation
	apiInstanceConversation := platformclientv2.NewConversationsApiWithConfig(config)

	reply, postError := getConversationId(*apiInstanceConversation, requestBody.Id)	// set request body object
	if postError != nil {
		response(w, postError.Error(), 400)
	}
	response(w, reply, 200)
}

func getConversationId(apiInstanceConversation platformclientv2.ConversationsApi, conversationId string) (reply string, error error) {
	data, response, err := apiInstanceConversation.GetConversation(conversationId)
	if err != nil {
		fmt.Printf("Error calling GetConversation: %v\n", err)
		return "", err
	} else {
		fmt.Printf("Response:  Success: %v  Status code: %v  Correlation ID: %v\n", response.IsSuccess, response.StatusCode, response.CorrelationID)
		json, err := json.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("response not json")
		}
		return string(json), nil
	}
}
