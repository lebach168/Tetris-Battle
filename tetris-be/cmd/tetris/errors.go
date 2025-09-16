package main

import "net/http"

func errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	e := envelope{"error": message}

	err := encodeComplexType(w, status, e, nil)
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(500)
	}

}

func serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {

	message := "the server encountered a problem and could not process your request"
	errorResponse(w, r, http.StatusInternalServerError, message)
}

func notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "resource not found"
	errorResponse(w, r, http.StatusNotFound, message)
}
func conflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	errorResponse(w, r, http.StatusConflict, message)
}
func badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	errorResponse(w, r, http.StatusBadRequest, err.Error())
}
func failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	errorResponse(w, r, http.StatusBadRequest, errors)
}
