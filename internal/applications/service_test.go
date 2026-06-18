package applications

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"jobscout/internal/database/sqlc"
	"jobscout/internal/httputil"
	"jobscout/internal/testdb"
)

func newApplicationTestService(t *testing.T) (*Service, *pgxpool.Pool) {
	t.Helper()

	pool := testdb.Open(t)
	return NewService(pool), pool
}

func createApplicationTestUser(t *testing.T, pool *pgxpool.Pool, email string) string {
	t.Helper()

	user, err := sqlc.New(pool).CreateUser(context.Background(), sqlc.CreateUserParams{
		Email:        email,
		PasswordHash: "hash",
	})
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}

	return user.ID.String()
}

func TestService_SaveApplication_ExternalAndDuplicate(t *testing.T) {
	svc, pool := newApplicationTestService(t)
	userID := createApplicationTestUser(t, pool, "external@example.com")

	req := SaveApplicationRequest{
		Source:          "remotive",
		ExternalID:      "job-1",
		Title:           "Backend Engineer",
		CompanyName:     "Remote Co",
		Category:        "Software Development",
		PublicationDate: appString("2026-06-01"),
		Status:          "Applied",
	}

	app, err := svc.SaveApplication(context.Background(), userID, req)
	if err != nil {
		t.Fatalf("save external application: %v", err)
	}
	if app.Source != "remotive" || app.ExternalID != "job-1" {
		t.Fatalf("unexpected saved application: %+v", app)
	}
	if app.Status != "Applied" {
		t.Fatalf("expected Applied, got %q", app.Status)
	}
	if app.PublicationDate == nil {
		t.Fatal("expected publication date")
	}

	_, err = svc.SaveApplication(context.Background(), userID, req)
	var conflictErr httputil.ConflictError
	if !errors.As(err, &conflictErr) {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func TestService_SaveApplication_ManualDefaults(t *testing.T) {
	svc, pool := newApplicationTestService(t)
	userID := createApplicationTestUser(t, pool, "manual@example.com")

	app, err := svc.SaveApplication(context.Background(), userID, SaveApplicationRequest{
		Title:       "Manual Job",
		CompanyName: "Manual Co",
	})
	if err != nil {
		t.Fatalf("save manual application: %v", err)
	}

	if app.Source != "manual" {
		t.Fatalf("expected manual source, got %q", app.Source)
	}
	if app.ExternalID == "" {
		t.Fatal("expected generated external id")
	}
	if app.Status != "Wishlist" {
		t.Fatalf("expected Wishlist, got %q", app.Status)
	}
}

func TestService_SaveApplicationValidation(t *testing.T) {
	svc, pool := newApplicationTestService(t)
	userID := createApplicationTestUser(t, pool, "validation@example.com")

	cases := []struct {
		name string
		req  SaveApplicationRequest
		want string
	}{
		{
			name: "external id required",
			req: SaveApplicationRequest{
				Source:      "remotive",
				Title:       "Backend Engineer",
				CompanyName: "Remote Co",
			},
			want: "external_id is required",
		},
		{
			name: "title required",
			req: SaveApplicationRequest{
				CompanyName: "Remote Co",
			},
			want: "title is required",
		},
		{
			name: "company required",
			req: SaveApplicationRequest{
				Title: "Backend Engineer",
			},
			want: "company_name is required",
		},
		{
			name: "invalid status",
			req: SaveApplicationRequest{
				Title:       "Backend Engineer",
				CompanyName: "Remote Co",
				Status:      "Unknown",
			},
			want: "invalid status",
		},
		{
			name: "invalid publication date",
			req: SaveApplicationRequest{
				Title:           "Backend Engineer",
				CompanyName:     "Remote Co",
				PublicationDate: appString("not-a-date"),
			},
			want: "invalid publication_date format",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.SaveApplication(context.Background(), userID, tc.req)
			var validationErr httputil.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("expected validation error, got %v", err)
			}
			if !strings.Contains(validationErr.Message, tc.want) {
				t.Fatalf("expected %q, got %q", tc.want, validationErr.Message)
			}
		})
	}
}

func TestService_ApplicationLifecycleAndUserScoping(t *testing.T) {
	svc, pool := newApplicationTestService(t)
	userID := createApplicationTestUser(t, pool, "owner@example.com")
	otherUserID := createApplicationTestUser(t, pool, "other@example.com")

	app, err := svc.SaveApplication(context.Background(), userID, SaveApplicationRequest{
		Source:      "remotive",
		ExternalID:  "job-lifecycle",
		Title:       "Backend Engineer",
		CompanyName: "Remote Co",
	})
	if err != nil {
		t.Fatalf("save application: %v", err)
	}

	otherApps, err := svc.ListApplications(context.Background(), otherUserID)
	if err != nil {
		t.Fatalf("list applications for other user: %v", err)
	}
	if len(otherApps) != 0 {
		t.Fatalf("expected no apps for other user, got %d", len(otherApps))
	}

	if _, err := svc.GetApplication(context.Background(), otherUserID, app.ID); !isNotFound(err) {
		t.Fatalf("expected not found for other user, got %v", err)
	}

	updated, err := svc.UpdateApplication(context.Background(), userID, app.ID, UpdateApplicationRequest{
		Title:       appString("Senior Backend Engineer"),
		CompanyName: appString("Remote Co Updated"),
	})
	if err != nil {
		t.Fatalf("update application: %v", err)
	}
	if updated.Title != "Senior Backend Engineer" || updated.CompanyName != "Remote Co Updated" {
		t.Fatalf("unexpected updated application: %+v", updated)
	}
	if updated.Status != "Wishlist" {
		t.Fatalf("status should not change through UpdateApplication, got %q", updated.Status)
	}

	statusUpdated, err := svc.UpdateApplicationStatus(context.Background(), userID, app.ID, "Interview")
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if statusUpdated.Status != "Interview" {
		t.Fatalf("expected Interview, got %q", statusUpdated.Status)
	}

	if _, err := svc.UpdateApplicationStatus(context.Background(), userID, app.ID, "Unknown"); !isValidation(err) {
		t.Fatalf("expected validation error, got %v", err)
	}

	if err := svc.DeleteApplication(context.Background(), otherUserID, app.ID); !isNotFound(err) {
		t.Fatalf("expected not found deleting other user's app, got %v", err)
	}
	if err := svc.DeleteApplication(context.Background(), userID, app.ID); err != nil {
		t.Fatalf("delete application: %v", err)
	}
	if _, err := svc.GetApplication(context.Background(), userID, app.ID); !isNotFound(err) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}

func TestService_ApplicationNotes(t *testing.T) {
	svc, pool := newApplicationTestService(t)
	userID := createApplicationTestUser(t, pool, "notes@example.com")
	otherUserID := createApplicationTestUser(t, pool, "notes-other@example.com")

	app, err := svc.SaveApplication(context.Background(), userID, SaveApplicationRequest{
		Title:       "Manual Job",
		CompanyName: "Manual Co",
	})
	if err != nil {
		t.Fatalf("save application: %v", err)
	}

	if _, err := svc.CreateApplicationNote(context.Background(), userID, app.ID, CreateNoteRequest{}); !isValidation(err) {
		t.Fatalf("expected empty body validation, got %v", err)
	}
	if _, err := svc.CreateApplicationNote(context.Background(), otherUserID, app.ID, CreateNoteRequest{Body: "private"}); !isNotFound(err) {
		t.Fatalf("expected not found for other user, got %v", err)
	}

	note, err := svc.CreateApplicationNote(context.Background(), userID, app.ID, CreateNoteRequest{Body: " Follow up next week "})
	if err != nil {
		t.Fatalf("create note: %v", err)
	}
	if note.Body != "Follow up next week" {
		t.Fatalf("expected trimmed body, got %q", note.Body)
	}

	notes, err := svc.ListApplicationNotes(context.Background(), userID, app.ID)
	if err != nil {
		t.Fatalf("list notes: %v", err)
	}
	if len(notes) != 1 || notes[0].ID != note.ID {
		t.Fatalf("unexpected notes: %+v", notes)
	}
	if _, err := svc.ListApplicationNotes(context.Background(), otherUserID, app.ID); !isNotFound(err) {
		t.Fatalf("expected not found listing other user's notes, got %v", err)
	}

	if err := svc.DeleteApplicationNote(context.Background(), otherUserID, note.ID); !isNotFound(err) {
		t.Fatalf("expected not found deleting other user's note, got %v", err)
	}
	if err := svc.DeleteApplicationNote(context.Background(), userID, note.ID); err != nil {
		t.Fatalf("delete note: %v", err)
	}
	if err := svc.DeleteApplicationNote(context.Background(), userID, note.ID); !isNotFound(err) {
		t.Fatalf("expected not found after deleting note, got %v", err)
	}
}

func appString(s string) *string {
	return &s
}

func isValidation(err error) bool {
	var validationErr httputil.ValidationError
	return errors.As(err, &validationErr)
}

func isNotFound(err error) bool {
	var notFoundErr httputil.NotFoundError
	return errors.As(err, &notFoundErr)
}
