package v1

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type responseWriterArgs struct {
	writer     http.ResponseWriter
	statusCode int

	// status specifies if the request is successful
	status  bool
	message string
}

// sendJSONResponse writes JSON-formatted response to client.
// If any error occurs while attempting to send JSON response,
// Response content-type default to text/html and status code is sent in header
func (app *App) sendJSONResponse(args *responseWriterArgs, data interface{}) {

	response := struct {
		Status  bool        `json:"status"`
		Message string      `json:"message"`
		Data    interface{} `json:"data,omitempty"`
	}{
		Status:  args.status,
		Message: args.message,
		Data:    data,
	}

	// Encode the data to JSON, returning the error if there was one.
	apiResponse, err := json.MarshalIndent(response, "", "\t")
	if err != nil {
		log.Printf("error encoding response to JSON: %v", err)
		return
	}

	args.writer.Header().Add("Content-Type", "application/json")
	_, err = args.writer.Write(apiResponse)
	if err != nil {
		args.writer.WriteHeader(500)
		log.Printf("error sending JSON response to client: %v", err)
	}
}

// sendJSONErrorResponse() method is a generic helper for sending JSON-formatted error
// messages to the client with a given status code.
func (app *App) sendJSONErrorResponse(w http.ResponseWriter, status int, message string, errors map[string]string) {

	response := struct {
		Status  bool              `json:"status"`
		Message string            `json:"message"`
		Errors  map[string]string `json:"errors,omitempty"`
	}{
		Status:  false,
		Message: message,
		Errors:  errors,
	}

	// Format the data to JSON
	apiResponse, err := json.MarshalIndent(response, "", "\t")
	if err != nil {
		log.Printf("error encoding response to JSON: %v", err)
		return
	}

	// set header values and send
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(apiResponse)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("error sending JSON response to client: %v", err)
	}
}

// serverErrorResponse sends a custom 500 internal server error to client.
func (app *App) sendServerErrorResponse(w http.ResponseWriter, err error) {
	log.Printf("internal error encountered: %v", err)
	message := "the server encountered an error and could not process your request"
	app.sendJSONErrorResponse(w, http.StatusInternalServerError, message, nil)
}

// sendNotFoundResponse sends a custom 404 not found status to client
func (app *App) sendNotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.sendJSONErrorResponse(w, http.StatusNotFound, message, nil)
}

// sendMethodNotAllowedResponse sends a 405 method not allowed to client.
func (app *App) sendMethodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.sendJSONErrorResponse(w, http.StatusMethodNotAllowed, message, nil)
}

// badRequestResponse method will be used to send a 400 Bad Request status code
// and JSON response to the client.
func (app *App) sendBadRequestResponse(w http.ResponseWriter, err error) {
	app.sendJSONErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
}

// sendFailedValidationResponse method sends a 422 Unprocessable Entity to client.
func (app *App) sendFailedValidationResponse(w http.ResponseWriter, errors map[string]string) {
	app.sendJSONErrorResponse(w, http.StatusUnprocessableEntity, "failed validation", errors)
}

// notValidString generically validates that text is not a valid string.
// notValidString can be further expanded with more validation logic
func notValidString(text string) bool {
	return text == ""
}
