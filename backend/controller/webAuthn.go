package controller

import (
	"api/config"
	"api/db"
	"api/helpers"
	"api/models"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	webauthn2 "github.com/go-webauthn/webauthn/webauthn"
	"net/http"
	"strconv"
)

// https://developers.google.com/codelabs/passkey-form-autofill

func GetUserByUrlParam(w http.ResponseWriter, r *http.Request) (user models.User, valid bool) {
	valid = true
	id, err := strconv.ParseUint(chi.URLParam(r, "userId"), 10, 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		valid = false
		return
	}

	result := db.DB.First(&user, id)
	if result.Error != nil {
		helpers.DBErrorHandling(result.Error, w)
		valid = false
	}

	return
}

func ExecAuthResponse(w http.ResponseWriter, r *http.Request, user models.User) error {
	// Get the session data stored from the function above
	sessionData := models.SessionData{
		UserId: user.WebAuthnID(),
	}
	result := db.DB.First(&sessionData)
	if result.Error == nil {
		result = db.DB.Delete(&sessionData)
	}
	if result.Error != nil {
		helpers.DBErrorHandling(result.Error, w)
		return errors.New("could not delete session data")
	}

	session := webauthn2.SessionData{
		Challenge:            sessionData.Challenge,
		UserID:               sessionData.UserId,
		AllowedCredentialIDs: sessionData.AllowedCredentialIds,
		Expires:              sessionData.Expires,
		UserVerification:     sessionData.UserVerification,
		Extensions:           sessionData.Extensions,
	}

	credential, err := config.WebAuthn.FinishLogin(user, session, r)
	if err != nil {
		http.Error(w, "Could not finish log in", http.StatusInternalServerError)
		return err
	}

	// If login was successful, update the credential object
	transports := make([]string, 0)
	for _, val := range credential.Transport {
		transports = append(transports, string(val))
	}

	dbCredential := models.Credential{
		WebauthnId:      credential.ID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		Transport:       transports,
		Flags: models.Flags{
			UserPresent:    credential.Flags.UserPresent,
			UserVerified:   credential.Flags.UserVerified,
			BackupEligible: credential.Flags.BackupEligible,
			BackupState:    credential.Flags.BackupState,
		},
		Authenticator: models.Authenticator{
			AAGUID:       credential.Authenticator.AAGUID,
			SignCount:    credential.Authenticator.SignCount,
			CloneWarning: credential.Authenticator.CloneWarning,
			Attachment:   string(credential.Authenticator.Attachment),
		},
	}

	result = db.DB.Save(&dbCredential)
	if result.Error != nil {
		helpers.DBErrorHandling(result.Error, w)
		return errors.New("could not save credentials")
	}
	return nil
}

// https://webauthn.guide/#registration
func RegisterRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, valid := GetUserByUrlParam(w, r)
	if !valid {
		return
	}

	options, session, err := config.WebAuthn.BeginRegistration(user)

	// store session data values
	sessionData := models.SessionData{
		Challenge:            session.Challenge,
		UserId:               session.UserID,
		AllowedCredentialIds: session.AllowedCredentialIDs,
		Expires:              session.Expires,
		UserVerification:     session.UserVerification,
		Extensions:           session.Extensions,
	}

	// handle errors
	if err != nil {
		http.Error(w, "Could not begin registration for provided user", http.StatusInternalServerError)
		return
	}

	// store sesion data
	result := db.DB.Create(&sessionData)
	if result.Error != nil {
		helpers.DBErrorHandling(result.Error, w)
		return
	}

	errJson := json.NewEncoder(w).Encode(&options.Response)
	if errJson != nil {
		http.Error(w, "Could not return results", http.StatusInternalServerError)
		return
	}
}

func RegisterResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, valid := GetUserByUrlParam(w, r)
	if !valid {
		return
	}

	// Get the session data stored from the function above
	sessionData := models.SessionData{
		UserId: user.WebAuthnID(),
	}
	result := db.DB.First(&sessionData)
	if result.Error == nil {
		result = db.DB.Delete(&sessionData)
	}
	if result.Error != nil {
		helpers.DBErrorHandling(result.Error, w)
		return
	}

	session := webauthn2.SessionData{
		Challenge:            sessionData.Challenge,
		UserID:               sessionData.UserId,
		AllowedCredentialIDs: sessionData.AllowedCredentialIds,
		Expires:              sessionData.Expires,
		UserVerification:     sessionData.UserVerification,
		Extensions:           sessionData.Extensions,
	}

	credential, err := config.WebAuthn.FinishRegistration(user, session, r)
	if err != nil {
		http.Error(w, "Could not finish registration for provided user", http.StatusInternalServerError)
		return
	}

	// save credentials
	transports := make([]string, 0)
	for _, val := range credential.Transport {
		transports = append(transports, string(val))
	}

	// Add public key to DB
	dbCredential := models.Credential{
		UserId:          user.Id,
		Roles:           nil, // TODO: use activeTokens table to find roles
		WebauthnId:      credential.ID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		Transport:       transports,
		Flags: models.Flags{
			UserPresent:    credential.Flags.UserPresent,
			UserVerified:   credential.Flags.UserVerified,
			BackupEligible: credential.Flags.BackupEligible,
			BackupState:    credential.Flags.BackupState,
		},
		Authenticator: models.Authenticator{
			AAGUID:       credential.Authenticator.AAGUID,
			SignCount:    credential.Authenticator.SignCount,
			CloneWarning: credential.Authenticator.CloneWarning,
			Attachment:   string(credential.Authenticator.Attachment),
		},
	}

	result = db.DB.Create(&dbCredential)
	if result.Error != nil {
		helpers.DBErrorHandling(result.Error, w)
		return
	}

	// TODO: delete token from activeTokens table

	errJson := json.NewEncoder(w).Encode("Registration Success")
	if errJson != nil {
		http.Error(w, "Could not return results", http.StatusInternalServerError)
		return
	}
}

func SigninRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Find the user
	user, valid := GetUserByUrlParam(w, r)
	if !valid {
		return
	}

	options, session, err := config.WebAuthn.BeginLogin(user)
	if err != nil {
		http.Error(w, "Could not being log in", http.StatusInternalServerError)
		return
	}

	// store the session values
	sessionData := models.SessionData{
		Challenge:            session.Challenge,
		UserId:               session.UserID,
		AllowedCredentialIds: session.AllowedCredentialIDs,
		Expires:              session.Expires,
		UserVerification:     session.UserVerification,
		Extensions:           session.Extensions,
	}

	// store sesion data
	result := db.DB.Create(&sessionData)
	if result.Error != nil {
		helpers.DBErrorHandling(result.Error, w)
		return
	}

	errJson := json.NewEncoder(w).Encode(&options.Response)
	if errJson != nil {
		http.Error(w, "Could not return results", http.StatusInternalServerError)
		return
	}
}

func SigninResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the user
	user, valid := GetUserByUrlParam(w, r)
	if !valid {
		return
	}

	err := ExecAuthResponse(w, r, user)
	if err != nil {
		return
	}

	errJson := json.NewEncoder(w).Encode("Login Success")
	if errJson != nil {
		http.Error(w, "Could not return results", http.StatusInternalServerError)
		return
	}
}