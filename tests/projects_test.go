package tests

import (
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"email-summary-tool/database/factories"
	"email-summary-tool/database/models"
	"email-summary-tool/pkg"
	"email-summary-tool/server/requests"

	"github.com/google/uuid"
)

func TestProjectList(t *testing.T) {
	project, err := factories.NewProject(t.Context(), TestServer.Db, TestServer.User.ID)
	if err != nil {
		t.Fatalf("create project fixture: %v", err)
	}

	otherUser := &models.User{
		Email: "other-project-user@example.com",
	}
	if _, err := TestServer.Db.NewInsert().Model(otherUser).Exec(t.Context()); err != nil {
		t.Fatalf("create other user fixture: %v", err)
	}
	otherProject, err := factories.NewProject(t.Context(), TestServer.Db, otherUser.ID)
	if err != nil {
		t.Fatalf("create other user's project fixture: %v", err)
	}

	request := pkg.Request{
		Method: http.MethodGet,
		Path:   "/projects",
	}

	cases := []pkg.TestCase{
		{
			TestName: "Can list projects",
			Request:  request,
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusOK,
				Body: map[string]any{
					"projects": []any{
						map[string]any{
							"id":          project.ID,
							"name":        project.Title,
							"description": project.Description,
							"status":      string(project.Status),
						},
					},
					"page":      1,
					"page_size": 10,
					"total":     1,
				},
			},
		},
		{
			TestName: "Cannot list projects belonging to another user",
			Request:  request,
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusOK,
				Body: map[string]any{
					"projects": []any{
						map[string]any{
							"id":          project.ID,
							"name":        project.Title,
							"description": project.Description,
							"status":      string(project.Status),
						},
					},
					"page":      1,
					"page_size": 10,
					"total":     1,
				},
				BodyDoesNotExist: map[string]any{
					"id": otherProject.ID,
				},
			},
		},
	}
	TestServer.RunAll(t, cases)
}

func TestProjectsGet(t *testing.T) {
	project, err := factories.NewProject(t.Context(), TestServer.Db, TestServer.User.ID)
	if err != nil {
		t.Fatalf("create project fixture: %v", err)
	}

	otherUser := &models.User{Email: "get-project-other-user@example.com"}
	if _, err := TestServer.Db.NewInsert().Model(otherUser).Exec(t.Context()); err != nil {
		t.Fatalf("create other user fixture: %v", err)
	}
	otherProject, err := factories.NewProject(t.Context(), TestServer.Db, otherUser.ID)
	if err != nil {
		t.Fatalf("create other user's project fixture: %v", err)
	}

	cases := []pkg.TestCase{
		{
			TestName: "Can get a project",
			Request: pkg.Request{
				Method: http.MethodGet,
				Path:   "/projects/" + project.ID,
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusOK,
				Body: map[string]any{
					"id":          project.ID,
					"name":        project.Title,
					"description": project.Description,
					"status":      string(project.Status),
				},
			},
		},
		{
			TestName: "Cannot get a project that does not exist",
			Request: pkg.Request{
				Method: http.MethodGet,
				Path:   "/projects/" + uuid.NewString(),
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusNotFound,
				Body:       "Project not found",
			},
		},
		{
			TestName: "Cannot get a project with an invalid ID",
			Request: pkg.Request{
				Method: http.MethodGet,
				Path:   "/projects/not-a-uuid",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusNotFound,
				Body:       "Project not found",
			},
		},
		{
			TestName: "Cannot get another user's project",
			Request: pkg.Request{
				Method: http.MethodGet,
				Path:   "/projects/" + otherProject.ID,
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusNotFound,
				Body:       "Project not found",
				BodyDoesNotExist: map[string]any{
					"id": otherProject.ID,
				},
			},
		},
		{
			TestName: "Cannot get a project without authentication",
			Request: pkg.Request{
				Method:          http.MethodGet,
				Path:            "/projects/" + project.ID,
				Unauthenticated: true,
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusUnauthorized,
				Body:       "Authentication required",
			},
		},
	}

	TestServer.RunAll(t, cases)
}

func TestProjectsCreate(t *testing.T) {

	request := pkg.Request{
		Method: http.MethodPost,
		Path:   "/projects",
	}

	cases := []pkg.TestCase{
		{
			TestName: "Can create a project",
			Request:  request,
			RequestBody: map[string]interface{}{
				"name":        "Daily summary",
				"description": "Summarise the daily inbox",
				"status":      "pending",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusCreated,
				Body: map[string]any{
					"name":        "Daily summary",
					"description": "Summarise the daily inbox",
					"status":      "pending",
				},
			},
		},
		{
			TestName: "Cannot create project invalid status",
			Request:  request,
			RequestBody: requests.CreateProjectRequest{
				Name:   "invalid status",
				Status: "invalid",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusBadRequest,
				Body:       "Invalid project status",
			},
		},
	}

	TestServer.RunAll(t, cases)
}

func TestProjectsUpdate(t *testing.T) {
	project, err := factories.NewProject(t.Context(), TestServer.Db, TestServer.User.ID)
	if err != nil {
		t.Fatalf("create project fixture: %v", err)
	}

	otherUser := &models.User{Email: "update-project-other-user@example.com"}
	if _, err := TestServer.Db.NewInsert().Model(otherUser).Exec(t.Context()); err != nil {
		t.Fatalf("create other user fixture: %v", err)
	}
	otherProject, err := factories.NewProject(t.Context(), TestServer.Db, otherUser.ID)
	if err != nil {
		t.Fatalf("create other user's project fixture: %v", err)
	}

	cases := []pkg.TestCase{
		{
			TestName: "Can update a project",
			Request: pkg.Request{
				Method: http.MethodPut,
				Path:   "/projects/" + project.ID,
			},
			RequestBody: requests.UpdateProjectRequest{
				Name:        "Updated project",
				Description: "Updated project description",
				Status:      "in_progress",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusOK,
				Body: map[string]any{
					"id":          project.ID,
					"name":        "Updated project",
					"description": "Updated project description",
					"status":      "in_progress",
				},
			},
		},
		{
			TestName: "Cannot update a project with an invalid status",
			Request: pkg.Request{
				Method: http.MethodPut,
				Path:   "/projects/" + project.ID,
			},
			RequestBody: requests.UpdateProjectRequest{
				Name:        "Invalid status",
				Description: "This update must be rejected",
				Status:      "invalid",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusBadRequest,
				Body:       "Invalid project status",
			},
		},
		{
			TestName: "Cannot update a project with malformed JSON",
			Request: pkg.Request{
				Method: http.MethodPut,
				Path:   "/projects/" + project.ID,
			},
			RequestBody: `{"name":`,
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusBadRequest,
				Body:       "Invalid request body",
			},
		},
		{
			TestName: "Cannot update a project with missing required fields",
			Request: pkg.Request{
				Method: http.MethodPut,
				Path:   "/projects/" + project.ID,
			},
			RequestBody: requests.UpdateProjectRequest{
				Name:        "Incomplete update",
				Description: " ",
				Status:      "pending",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusBadRequest,
				Body:       "Invalid request body",
			},
		},
		{
			TestName: "Cannot update a project that does not exist",
			Request: pkg.Request{
				Method: http.MethodPut,
				Path:   "/projects/" + uuid.NewString(),
			},
			RequestBody: requests.UpdateProjectRequest{
				Name:        "Missing project",
				Description: "This project does not exist",
				Status:      "pending",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusNotFound,
				Body:       "Project not found",
			},
		},
		{
			TestName: "Cannot update a project with an invalid ID",
			Request: pkg.Request{
				Method: http.MethodPut,
				Path:   "/projects/not-a-uuid",
			},
			RequestBody: requests.UpdateProjectRequest{
				Name:        "Invalid project ID",
				Description: "This project ID is invalid",
				Status:      "pending",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusNotFound,
				Body:       "Project not found",
			},
		},
		{
			TestName: "Cannot update another user's project",
			Request: pkg.Request{
				Method: http.MethodPut,
				Path:   "/projects/" + otherProject.ID,
			},
			RequestBody: requests.UpdateProjectRequest{
				Name:        "Forbidden update",
				Description: "This project belongs to another user",
				Status:      "archived",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusNotFound,
				Body:       "Project not found",
			},
		},
		{
			TestName: "Cannot update a project without authentication",
			Request: pkg.Request{
				Method:          http.MethodPut,
				Path:            "/projects/" + project.ID,
				Unauthenticated: true,
			},
			RequestBody: requests.UpdateProjectRequest{
				Name:        "Unauthenticated update",
				Description: "This update must be rejected",
				Status:      "completed",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusUnauthorized,
				Body:       "Authentication required",
			},
		},
	}

	TestServer.RunAll(t, cases)

	persistedProject := &models.Project{}
	if err := TestServer.Db.NewSelect().Model(persistedProject).Where("id = ?", project.ID).Scan(t.Context()); err != nil {
		t.Fatalf("get updated project: %v", err)
	}
	if persistedProject.Title != "Updated project" ||
		persistedProject.Description != "Updated project description" ||
		persistedProject.Status != "in_progress" {
		t.Errorf("updated project was not persisted: %#v", persistedProject)
	}

	persistedOtherProject := &models.Project{}
	if err := TestServer.Db.NewSelect().Model(persistedOtherProject).Where("id = ?", otherProject.ID).Scan(t.Context()); err != nil {
		t.Fatalf("get other user's project: %v", err)
	}
	if persistedOtherProject.Title != otherProject.Title ||
		persistedOtherProject.Description != otherProject.Description ||
		persistedOtherProject.Status != otherProject.Status {
		t.Errorf("other user's project was updated: %#v", persistedOtherProject)
	}
}

func TestProjectsDelete(t *testing.T) {
	project, err := factories.NewProject(t.Context(), TestServer.Db, TestServer.User.ID)
	if err != nil {
		t.Fatalf("create project fixture: %v", err)
	}

	otherUser := &models.User{Email: "delete-project-other-user@example.com"}
	if _, err := TestServer.Db.NewInsert().Model(otherUser).Exec(t.Context()); err != nil {
		t.Fatalf("create other user fixture: %v", err)
	}
	otherProject, err := factories.NewProject(t.Context(), TestServer.Db, otherUser.ID)
	if err != nil {
		t.Fatalf("create other user's project fixture: %v", err)
	}

	cases := []pkg.TestCase{
		{
			TestName: "Can delete a project",
			Request: pkg.Request{
				Method: http.MethodDelete,
				Path:   "/projects/" + project.ID,
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusNoContent,
				Body:       "",
			},
		},
		{
			TestName: "Cannot delete a project that does not exist",
			Request: pkg.Request{
				Method: http.MethodDelete,
				Path:   "/projects/" + project.ID,
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusNotFound,
				Body:       "Project not found",
			},
		},
		{
			TestName: "Cannot delete a project with an invalid ID",
			Request: pkg.Request{
				Method: http.MethodDelete,
				Path:   "/projects/not-a-uuid",
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusNotFound,
				Body:       "Project not found",
			},
		},
		{
			TestName: "Cannot delete another user's project",
			Request: pkg.Request{
				Method: http.MethodDelete,
				Path:   "/projects/" + otherProject.ID,
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusNotFound,
				Body:       "Project not found",
			},
		},
		{
			TestName: "Cannot delete a project without authentication",
			Request: pkg.Request{
				Method:          http.MethodDelete,
				Path:            "/projects/" + otherProject.ID,
				Unauthenticated: true,
			},
			ExpectedResult: pkg.ExpectedResult{
				StatusCode: http.StatusUnauthorized,
				Body:       "Authentication required",
			},
		},
	}

	TestServer.RunAll(t, cases)

	deletedProject := &models.Project{}
	err = TestServer.Db.NewSelect().Model(deletedProject).Where("id = ?", project.ID).Scan(t.Context())
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("get deleted project error = %v, want %v", err, sql.ErrNoRows)
	}

	persistedOtherProject := &models.Project{}
	if err := TestServer.Db.NewSelect().Model(persistedOtherProject).Where("id = ?", otherProject.ID).Scan(t.Context()); err != nil {
		t.Fatalf("get other user's project after delete request: %v", err)
	}
}
