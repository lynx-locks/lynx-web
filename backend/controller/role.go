package controller

import (
	"api/db"
	"api/helpers"
	"api/models"
	"encoding/json"
	"net/http"
)

func GetAllRoles(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err, roles := helpers.GetAllTable(w, []models.Role{})
	if err != nil {
		return
	}
	helpers.JsonWriter(w, roles)
}

func UpdateRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err, rId := helpers.ParseInt(w, r, "roleId")
	if err != nil {
		return
	}
	err, role := helpers.GetFirstTable(w, models.Role{}, models.Common{Id: rId})
	if err != nil {
		return
	}
	err, role = helpers.UpdateSpecifiedParams(w, r, &role, role.Common, &role.Common)
	if err != nil {
		return
	}
	helpers.JsonWriter(w, role)

}

func CreateRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var role models.Role
	err := json.NewDecoder(r.Body).Decode(&role)
	if err != nil {
		http.Error(w, "400 malformed request", http.StatusBadRequest)
	}
	// essentially reset if the user inputted anything to common as it should not be editable
	role.Common = models.Common{}
	err, role = helpers.CreateNewRecord(w, role)
	if err != nil {
		return
	}
	helpers.JsonWriter(w, role)
}

func DeleteRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err, rId := helpers.ParseInt(w, r, "roleId")
	if err != nil {
		return
	}
	err = helpers.DeleteByPk(w, models.Role{}, rId)
	if err != nil {
		return
	}
	err = db.DB.Unscoped().Model(&models.Role{Common: models.Common{Id: rId}}).Association("Doors").Unscoped().Clear()
	if err != nil {
		http.Error(w, "500 unable to delete association/s", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)
}

func ReplaceDoorAssociation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err, rId := helpers.ParseInt(w, r, "roleId")
	if err != nil {
		return
	}
	doorIDs := []uint{}
	err = json.NewDecoder(r.Body).Decode(&doorIDs)
	if err != nil {
		http.Error(w, "400 malformed doorIDs", http.StatusBadRequest)
		return
	}
	doors := []models.Door{}

	if len(doorIDs) == 0 {
		http.Error(w, "400 missing door ids", http.StatusBadRequest)
		return
	}
	results := db.DB.Where(&doorIDs).Find(&doors)

	if results.RowsAffected != int64(len(doorIDs)) {
		http.Error(w, "400 one or more invalid door id/s", http.StatusBadRequest)
		return
	}

	err, role := helpers.GetFirstTable(w, models.Role{}, models.Role{Common: models.Common{Id: uint(rId)}})
	if err != nil {
		helpers.DBErrorHandling(err, w)
		return
	}

	err = db.DB.Model(&role).Association("Doors").Replace(&doors)
	if err != nil {
		helpers.DBErrorHandling(err, w)
		return
	}
	helpers.JsonWriter(w, &role)
}

func GetDoorAssociations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err, rId := helpers.ParseInt(w, r, "roleId")
	if err != nil {
		return
	}
	role := models.Role{Common: models.Common{Id: rId}}
	doors := []models.Door{}
	err = db.DB.Model(&role).Association("Doors").Find(&doors)
	if err != nil {
		helpers.DBErrorHandling(err, w)
		return
	}
	helpers.JsonWriter(w, &doors)
}
