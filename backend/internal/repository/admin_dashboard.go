package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/kareelio/backend/internal/model"
)

type AdminDashboardRepository struct {
	db *pgxpool.Pool
}

func NewAdminDashboardRepository(db *pgxpool.Pool) *AdminDashboardRepository {
	return &AdminDashboardRepository{db: db}
}

func (r *AdminDashboardRepository) GetDashboard(ctx context.Context) (*model.AdminDashboard, error) {
	dash := &model.AdminDashboard{
		ByStatus:   make(map[string]int),
		BySource:   make(map[string]int),
		ByRemote:   make(map[string]int),
		ByPriority: make(map[string]int),
	}

	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 'user'`).Scan(&dash.Users.Total)
	if err != nil {
		return nil, fmt.Errorf("users total: %w", err)
	}

	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 'user' AND is_active = true AND email_verified_at IS NOT NULL`).Scan(&dash.Users.Active)
	if err != nil {
		return nil, fmt.Errorf("users active: %w", err)
	}

	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 'user' AND email_verified_at IS NULL`).Scan(&dash.Users.Unverified)
	if err != nil {
		return nil, fmt.Errorf("users unverified: %w", err)
	}

	dash.Users.Disabled = dash.Users.Total - dash.Users.Active - dash.Users.Unverified

	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM job_applications`).Scan(&dash.Applications.Total)
	if err != nil {
		return nil, fmt.Errorf("apps total: %w", err)
	}

	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM job_applications WHERE created_at >= NOW() - INTERVAL '7 days'`).Scan(&dash.Applications.CreatedLast7Days)
	if err != nil {
		return nil, fmt.Errorf("apps 7d: %w", err)
	}

	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM job_applications WHERE created_at >= NOW() - INTERVAL '30 days'`).Scan(&dash.Applications.CreatedLast30Days)
	if err != nil {
		return nil, fmt.Errorf("apps 30d: %w", err)
	}

	if dash.Users.Active > 0 {
		dash.Applications.AveragePerActiveUser = float64(dash.Applications.Total) / float64(dash.Users.Active)
	}

	appliedCount := 0
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM job_applications WHERE status != 'draft'`).Scan(&appliedCount)
	if err != nil {
		return nil, fmt.Errorf("applied count: %w", err)
	}

	if appliedCount > 0 {
		var respondedCount, interviewCount, testCount, offerCount int

		_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM job_applications WHERE response_received = true AND status != 'draft'`).Scan(&respondedCount)
		_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM job_applications WHERE first_contact_date IS NOT NULL`).Scan(&interviewCount)
		_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM job_applications WHERE has_test = true`).Scan(&testCount)
		_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM job_applications WHERE offer_received = true`).Scan(&offerCount)

		dash.Funnels.ResponseRate = float64(respondedCount) / float64(appliedCount) * 100
		dash.Funnels.InterviewRate = float64(interviewCount) / float64(appliedCount) * 100
		dash.Funnels.TestRate = float64(testCount) / float64(appliedCount) * 100
		dash.Funnels.OfferRate = float64(offerCount) / float64(appliedCount) * 100
	}

	statusRows, err := r.db.Query(ctx, `SELECT status, COUNT(*) FROM job_applications GROUP BY status`)
	if err == nil {
		defer statusRows.Close()
		for statusRows.Next() {
			var status string
			var count int
			if statusRows.Scan(&status, &count) == nil {
				dash.ByStatus[status] = count
			}
		}
	}

	sourceRows, err := r.db.Query(ctx, `SELECT source, COUNT(*) FROM job_applications GROUP BY source`)
	if err == nil {
		defer sourceRows.Close()
		for sourceRows.Next() {
			var source string
			var count int
			if sourceRows.Scan(&source, &count) == nil {
				dash.BySource[source] = count
			}
		}
	}

	remoteRows, err := r.db.Query(ctx, `SELECT remote, COUNT(*) FROM job_applications GROUP BY remote`)
	if err == nil {
		defer remoteRows.Close()
		for remoteRows.Next() {
			var remote string
			var count int
			if remoteRows.Scan(&remote, &count) == nil {
				dash.ByRemote[remote] = count
			}
		}
	}

	priorityRows, err := r.db.Query(ctx, `SELECT priority, COUNT(*) FROM job_applications GROUP BY priority`)
	if err == nil {
		defer priorityRows.Close()
		for priorityRows.Next() {
			var priority string
			var count int
			if priorityRows.Scan(&priority, &count) == nil {
				dash.ByPriority[priority] = count
			}
		}
	}

	return dash, nil
}
