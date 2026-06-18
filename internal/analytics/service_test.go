package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"jobscout/internal/database/sqlc"
	"jobscout/internal/testdb"
)

func newAnalyticsTestService(t *testing.T) (*Service, *pgxpool.Pool) {
	t.Helper()

	pool := testdb.Open(t)
	return NewService(pool), pool
}

func createAnalyticsTestUser(t *testing.T, pool *pgxpool.Pool, email string) sqlc.User {
	t.Helper()

	user, err := sqlc.New(pool).CreateUser(context.Background(), sqlc.CreateUserParams{
		Email:        email,
		PasswordHash: "hash",
	})
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}

	return user
}

func createAnalyticsTestApplication(t *testing.T, pool *pgxpool.Pool, userID pgtype.UUID, arg analyticsApplicationSeed) {
	t.Helper()

	app, err := sqlc.New(pool).CreateApplication(context.Background(), sqlc.CreateApplicationParams{
		UserID:      userID,
		Source:      arg.source,
		ExternalID:  arg.externalID,
		Title:       arg.title,
		CompanyName: arg.company,
		Category:    arg.category,
		Status:      arg.status,
	})
	if err != nil {
		t.Fatalf("create test application: %v", err)
	}

	if _, err := pool.Exec(context.Background(), `UPDATE applications SET created_at = $1 WHERE id = $2`, arg.createdAt, app.ID); err != nil {
		t.Fatalf("set test application created_at: %v", err)
	}
}

func TestService_SummaryEmpty(t *testing.T) {
	svc, pool := newAnalyticsTestService(t)
	user := createAnalyticsTestUser(t, pool, "analytics-empty@example.com")

	summary, err := svc.Summary(context.Background(), user.ID.String())
	if err != nil {
		t.Fatalf("summary failed: %v", err)
	}

	if summary.TotalSavedJobs != 0 {
		t.Fatalf("expected 0 saved jobs, got %d", summary.TotalSavedJobs)
	}
	if len(summary.ByStatus) != 0 || len(summary.BySource) != 0 || len(summary.ByCategory) != 0 {
		t.Fatalf("expected empty maps, got %+v", summary)
	}
	if len(summary.SavedPerMonth) != 0 {
		t.Fatalf("expected no month counts, got %+v", summary.SavedPerMonth)
	}
}

func TestService_SummaryAggregatesByUser(t *testing.T) {
	svc, pool := newAnalyticsTestService(t)
	user := createAnalyticsTestUser(t, pool, "analytics@example.com")
	otherUser := createAnalyticsTestUser(t, pool, "analytics-other@example.com")

	createAnalyticsTestApplication(t, pool, user.ID, analyticsApplicationSeed{
		source:     "remotive",
		externalID: "job-1",
		title:      "Backend Engineer",
		company:    "Remote Co",
		category:   "Software Development",
		status:     "Wishlist",
		createdAt:  time.Date(2026, 5, 10, 10, 0, 0, 0, time.UTC),
	})
	createAnalyticsTestApplication(t, pool, user.ID, analyticsApplicationSeed{
		source:     "remotive",
		externalID: "job-2",
		title:      "Frontend Engineer",
		company:    "Remote Co",
		category:   "Software Development",
		status:     "Applied",
		createdAt:  time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC),
	})
	createAnalyticsTestApplication(t, pool, user.ID, analyticsApplicationSeed{
		source:     "manual",
		externalID: "manual-1",
		title:      "Referral Role",
		company:    "Referral Co",
		category:   "",
		status:     "Applied",
		createdAt:  time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC),
	})
	createAnalyticsTestApplication(t, pool, otherUser.ID, analyticsApplicationSeed{
		source:     "manual",
		externalID: "other-1",
		title:      "Other Role",
		company:    "Other Co",
		category:   "Data",
		status:     "Offer",
		createdAt:  time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC),
	})

	summary, err := svc.Summary(context.Background(), user.ID.String())
	if err != nil {
		t.Fatalf("summary failed: %v", err)
	}

	if summary.TotalSavedJobs != 3 {
		t.Fatalf("expected 3 saved jobs, got %d", summary.TotalSavedJobs)
	}
	if summary.ByStatus["Applied"] != 2 || summary.ByStatus["Wishlist"] != 1 {
		t.Fatalf("unexpected status counts: %+v", summary.ByStatus)
	}
	if summary.BySource["remotive"] != 2 || summary.BySource["manual"] != 1 {
		t.Fatalf("unexpected source counts: %+v", summary.BySource)
	}
	if summary.ByCategory["Software Development"] != 2 || summary.ByCategory["Uncategorized"] != 1 {
		t.Fatalf("unexpected category counts: %+v", summary.ByCategory)
	}
	if len(summary.SavedPerMonth) != 2 {
		t.Fatalf("expected 2 month counts, got %+v", summary.SavedPerMonth)
	}
	if summary.SavedPerMonth[0] != (MonthCount{Month: "2026-05", Count: 1}) {
		t.Fatalf("unexpected first month count: %+v", summary.SavedPerMonth[0])
	}
	if summary.SavedPerMonth[1] != (MonthCount{Month: "2026-06", Count: 2}) {
		t.Fatalf("unexpected second month count: %+v", summary.SavedPerMonth[1])
	}
}

type analyticsApplicationSeed struct {
	source     string
	externalID string
	title      string
	company    string
	category   string
	status     string
	createdAt  time.Time
}
