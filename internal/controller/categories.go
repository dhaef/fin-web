package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"fin-web/internal/model"
)

type CategoriesPage struct {
	Categories []model.Category
}

type CategoryPage struct {
	Category model.Category
	Type     string
}

func (c *Controller) categories(w http.ResponseWriter, r *http.Request) error {
	categories, err := model.GetCategories(c.db)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error getting categories: " + err.Error(),
		}
	}

	err = renderTemplate(w, Base[CategoriesPage]{
		Data: CategoriesPage{
			Categories: categories,
		},
	}, "layout", []string{"categories/categories.html", "layout.html"})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return nil
}

func (c *Controller) category(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")

	cat, err := model.GetCategory(c.db, id)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error getting category: " + err.Error(),
		}
	}

	err = renderTemplate(w, Base[CategoryPage]{
		Data: CategoryPage{
			Category: cat,
			Type:     "edit",
		},
	}, "layout", []string{"categories/category.html", "layout.html"})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return nil
}

func (c *Controller) newCategory(w http.ResponseWriter, r *http.Request) error {
	err := renderTemplate(w, Base[CategoryPage]{
		Data: CategoryPage{
			Type: "create",
		},
	}, "layout", []string{"categories/category.html", "layout.html"})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return nil
}

type CategoryValueFormItem struct {
	ID    *string `json:"id"`
	Value string  `json:"value"`
}

func (c *Controller) createCategory(w http.ResponseWriter, r *http.Request) error {
	labelStr := r.FormValue("label")
	priorityStr := r.FormValue("priority")
	typeStr := r.FormValue("type")
	isIgnoredStr := r.FormValue("is_ignored")
	valuesStr := r.FormValue("values")

	var values []CategoryValueFormItem
	err := json.Unmarshal([]byte(valuesStr), &values)
	if err != nil {
		return APIError{
			Status:  http.StatusBadRequest,
			Message: "error unmarshalling values: " + err.Error(),
		}
	}

	errs := map[string]string{}

	if labelStr == "" {
		errs["label"] = "label can not be empty"
	}

	if typeStr == "" {
		errs["type"] = "type can not be empty"
	}

	var priority int
	if priorityStr == "" {
		errs["priority"] = "priority can not be empty"
	} else {
		priority, err = strconv.Atoi(priorityStr)
		if err != nil {
			errs["priority"] = err.Error()
		}
	}

	for _, v := range values {
		if v.Value == "" {
			errs["values"] = "value can not be empty"
			break
		}
	}

	var isIgnored bool
	isIgnored, err = strconv.ParseBool(isIgnoredStr)
	if err != nil {
		errs["is_ignored"] = err.Error()
	}

	if len(errs) != 0 {
		return encode(w, r, http.StatusBadRequest, map[string]any{
			"errs": errs,
		})
	}

	ID, err := model.CreateCategory(c.db, labelStr, priority, typeStr, isIgnored)
	if err != nil {
		if err.Error() == "Error: This value already exists in the table." {
			return encode(w, r, http.StatusBadRequest, map[string]any{
				"errs": map[string]string{
					"priority": "This priority is already taken",
				},
			})
		}

		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error creating category: " + err.Error(),
		}
	}

	for _, v := range values {
		_, err = model.CreateCategoryValue(c.db, ID, v.Value)
		if err != nil {
			return APIError{
				Status:  http.StatusInternalServerError,
				Message: "error creating category value: " + err.Error(),
			}
		}
	}

	return encode(w, r, http.StatusCreated, map[string]string{
		"redirect": "/categories/" + strconv.Itoa(ID),
	})
}

func (c *Controller) updateCategory(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	labelStr := r.FormValue("label")
	priorityStr := r.FormValue("priority")
	valuesStr := r.FormValue("values")
	typeStr := r.FormValue("category_type")
	isIgnoredStr := r.FormValue("is_ignored")

	var values []CategoryValueFormItem
	err := json.Unmarshal([]byte(valuesStr), &values)
	if err != nil {
		return APIError{
			Status:  http.StatusBadRequest,
			Message: "error unmarshalling values: " + err.Error(),
		}
	}

	errs := map[string]string{}

	if labelStr == "" {
		errs["label"] = "label can not be empty"
	}

	if typeStr == "" {
		errs["type"] = "type can not be empty"
	}

	var priority int
	if priorityStr == "" {
		errs["priority"] = "priority can not be empty"
	} else {
		priority, err = strconv.Atoi(priorityStr)
		if err != nil {
			errs["priority"] = err.Error()
		}
	}

	for _, v := range values {
		if v.Value == "" {
			errs["values"] = "value can not be empty"
			break
		}
	}

	var isIgnored bool
	isIgnored, err = strconv.ParseBool(isIgnoredStr)
	if err != nil {
		errs["is_ignored"] = err.Error()
	}

	if len(errs) != 0 {
		return encode(w, r, http.StatusBadRequest, map[string]any{
			"errs": errs,
		})
	}

	currCategory, err := model.GetCategory(c.db, id)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error getting category: " + err.Error(),
		}
	}

	err = model.UpdateCategory(c.db, id, model.UpdateCategoryParams{
		Label:        &labelStr,
		Priority:     &priority,
		CategoryType: &typeStr,
		IsIgnored:    &isIgnored,
	})
	if err != nil {
		if err.Error() == "Error: This value already exists in another row." {
			return encode(w, r, http.StatusBadRequest, map[string]any{
				"errs": map[string]string{
					"priority": "This priority is already taken",
				},
			})
		}

		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error updating category: " + err.Error(),
		}
	}

	valueMap := map[string]CategoryValueFormItem{}
	valuesToCreate := []string{}
	valuesToUpdate := []CategoryValueFormItem{}
	valuesToDelete := []model.CategoryValue{}

	for _, v := range values {
		if v.ID == nil {
			valuesToCreate = append(valuesToCreate, v.Value)
			continue
		}
		valueMap[*v.ID] = v

		// check if it needs updating
		for _, i := range currCategory.Values {
			if i.ID.Valid && strconv.Itoa(int(i.ID.Int64)) == *v.ID && i.Value.String != v.Value {
				valuesToUpdate = append(valuesToUpdate, v)
			}
		}
	}

	for _, v := range currCategory.Values {
		_, ok := valueMap[strconv.Itoa(int(v.ID.Int64))]
		if !ok {
			valuesToDelete = append(valuesToDelete, v)
		}
	}

	for _, v := range valuesToCreate {
		_, err = model.CreateCategoryValue(c.db, currCategory.ID, v)
		if err != nil {
			return APIError{
				Status:  http.StatusInternalServerError,
				Message: "error creating category value: " + err.Error(),
			}
		}
	}

	for _, v := range valuesToUpdate {
		err = model.UpdateCategoryValue(c.db, *v.ID, model.UpdateCategoryValueParams{
			Value: &v.Value,
		})
		if err != nil {
			return APIError{
				Status:  http.StatusInternalServerError,
				Message: "error updating category value: " + err.Error(),
			}
		}
	}

	for _, v := range valuesToDelete {
		err = model.DeleteCategoryValue(c.db, int(v.ID.Int64))
		if err != nil {
			return APIError{
				Status:  http.StatusInternalServerError,
				Message: "error deleting category value: " + err.Error(),
			}
		}
	}

	return encode(w, r, http.StatusCreated, map[string]string{
		"redirect": "/categories/" + strconv.Itoa(currCategory.ID),
	})
}

func (c *Controller) deleteCategory(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")

	// get and delete category values
	cat, err := model.GetCategory(c.db, id)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error getting category: " + err.Error(),
		}
	}

	for _, val := range cat.Values {
		if val.ID.Valid {
			err = model.DeleteCategoryValue(c.db, int(val.ID.Int64))
			if err != nil {
				return APIError{
					Status:  http.StatusInternalServerError,
					Message: "error deleting category value: " + err.Error(),
				}
			}
		}
	}

	err = model.DeleteCategory(c.db, id)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error deleting category: " + err.Error(),
		}
	}

	http.Redirect(w, r, "/categories", http.StatusSeeOther)
	return nil
}
