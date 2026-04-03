package reports

import (
	"context"
	"fmt"

	"pension-manager/internal/db"
)

// Service manages contribution reporting
type Service struct {
	db *db.DB
}

// NewService creates a new reports service
func NewService(db *db.DB) *Service {
	return &Service{db: db}
}

// ContributionBreakdown holds employee/employer breakdown
type ContributionBreakdown struct {
	SchemeID      string `json:"scheme_id"`
	SchemeName    string `json:"scheme_name"`
	Period        string `json:"period"`
	EmployeeCount int    `json:"employee_count"`
	EmployeeTotal int64  `json:"employee_total"`
	EmployerTotal int64  `json:"employer_total"`
	AVCTotal      int64  `json:"avc_total"`
	GrandTotal    int64  `json:"grand_total"`
}

// GetEmployeeEmployerBreakdown returns contributions broken down by employee and employer
func (s *Service) GetEmployeeEmployerBreakdown(ctx context.Context, schemeID string, year int) ([]ContributionBreakdown, error) {
	query := `
		SELECT s.id, s.name, TO_CHAR(c.period, 'YYYY-MM') as period,
		       COUNT(DISTINCT c.member_id) as employee_count,
		       COALESCE(SUM(c.employee_amount), 0) as employee_total,
		       COALESCE(SUM(c.employer_amount), 0) as employer_total,
		       COALESCE(SUM(c.avc_amount), 0) as avc_total,
		       COALESCE(SUM(c.total_amount), 0) as grand_total
		FROM contributions c
		JOIN schemes s ON s.id = c.scheme_id
		WHERE c.scheme_id = $1 AND EXTRACT(YEAR FROM c.period) = $2
		GROUP BY s.id, s.name, TO_CHAR(c.period, 'YYYY-MM')
		ORDER BY period
	`
	rows, err := s.db.QueryContext(ctx, query, schemeID, year)
	if err != nil {
		return nil, fmt.Errorf("query breakdown: %w", err)
	}
	defer rows.Close()

	var breakdowns []ContributionBreakdown
	for rows.Next() {
		var b ContributionBreakdown
		if err := rows.Scan(&b.SchemeID, &b.SchemeName, &b.Period, &b.EmployeeCount,
			&b.EmployeeTotal, &b.EmployerTotal, &b.AVCTotal, &b.GrandTotal); err != nil {
			return nil, fmt.Errorf("scan breakdown: %w", err)
		}
		breakdowns = append(breakdowns, b)
	}
	return breakdowns, rows.Err()
}

// YTDContribution holds year-to-date contribution totals
type YTDContribution struct {
	MemberID    string `json:"member_id"`
	MemberNo    string `json:"member_no"`
	FullName    string `json:"full_name"`
	EmployeeYTD int64  `json:"employee_ytd"`
	EmployerYTD int64  `json:"employer_ytd"`
	AVCYTD      int64  `json:"avc_ytd"`
	TotalYTD    int64  `json:"total_ytd"`
}

// GetYTDContributions returns YTD contributions per member
func (s *Service) GetYTDContributions(ctx context.Context, schemeID string, year int) ([]YTDContribution, error) {
	query := `
		SELECT m.id, m.member_no, m.first_name || ' ' || m.last_name,
		       COALESCE(SUM(c.employee_amount), 0) as employee_ytd,
		       COALESCE(SUM(c.employer_amount), 0) as employer_ytd,
		       COALESCE(SUM(c.avc_amount), 0) as avc_ytd,
		       COALESCE(SUM(c.total_amount), 0) as total_ytd
		FROM members m
		LEFT JOIN contributions c ON c.member_id = m.id AND EXTRACT(YEAR FROM c.period) = $1
		WHERE m.scheme_id = $2
		GROUP BY m.id, m.member_no, m.first_name, m.last_name
		ORDER BY total_ytd DESC
	`
	rows, err := s.db.QueryContext(ctx, query, year, schemeID)
	if err != nil {
		return nil, fmt.Errorf("query YTD: %w", err)
	}
	defer rows.Close()

	var ytds []YTDContribution
	for rows.Next() {
		var y YTDContribution
		if err := rows.Scan(&y.MemberID, &y.MemberNo, &y.FullName, &y.EmployeeYTD,
			&y.EmployerYTD, &y.AVCYTD, &y.TotalYTD); err != nil {
			return nil, fmt.Errorf("scan YTD: %w", err)
		}
		ytds = append(ytds, y)
	}
	return ytds, rows.Err()
}

// CumulativeContribution holds cumulative contribution totals
type CumulativeContribution struct {
	MemberID           string `json:"member_id"`
	MemberNo           string `json:"member_no"`
	FullName           string `json:"full_name"`
	EmployeeCumulative int64  `json:"employee_cumulative"`
	EmployerCumulative int64  `json:"employer_cumulative"`
	AVCCumulative      int64  `json:"avc_cumulative"`
	TotalCumulative    int64  `json:"total_cumulative"`
}

// GetCumulativeContributions returns cumulative contributions per member
func (s *Service) GetCumulativeContributions(ctx context.Context, schemeID string) ([]CumulativeContribution, error) {
	query := `
		SELECT m.id, m.member_no, m.first_name || ' ' || m.last_name,
		       COALESCE(SUM(c.employee_amount), 0) as employee_cumulative,
		       COALESCE(SUM(c.employer_amount), 0) as employer_cumulative,
		       COALESCE(SUM(c.avc_amount), 0) as avc_cumulative,
		       COALESCE(SUM(c.total_amount), 0) as total_cumulative
		FROM members m
		LEFT JOIN contributions c ON c.member_id = m.id
		WHERE m.scheme_id = $1
		GROUP BY m.id, m.member_no, m.first_name, m.last_name
		ORDER BY total_cumulative DESC
	`
	rows, err := s.db.QueryContext(ctx, query, schemeID)
	if err != nil {
		return nil, fmt.Errorf("query cumulative: %w", err)
	}
	defer rows.Close()

	var cumulatives []CumulativeContribution
	for rows.Next() {
		var c CumulativeContribution
		if err := rows.Scan(&c.MemberID, &c.MemberNo, &c.FullName, &c.EmployeeCumulative,
			&c.EmployerCumulative, &c.AVCCumulative, &c.TotalCumulative); err != nil {
			return nil, fmt.Errorf("scan cumulative: %w", err)
		}
		cumulatives = append(cumulatives, c)
	}
	return cumulatives, rows.Err()
}

// RegisteredVsUnregistered holds registered vs unregistered contribution data
type RegisteredVsUnregistered struct {
	MemberID          string `json:"member_id"`
	MemberNo          string `json:"member_no"`
	FullName          string `json:"full_name"`
	RegisteredTotal   int64  `json:"registered_total"`
	UnregisteredTotal int64  `json:"unregistered_total"`
	Year              int    `json:"year"`
}

// GetRegisteredVsUnregistered returns registered vs unregistered contributions
func (s *Service) GetRegisteredVsUnregistered(ctx context.Context, schemeID string, year int) ([]RegisteredVsUnregistered, error) {
	query := `
		SELECT m.id, m.member_no, m.first_name || ' ' || m.last_name,
		       COALESCE(SUM(CASE WHEN c.registered THEN c.total_amount ELSE 0 END), 0) as registered_total,
		       COALESCE(SUM(CASE WHEN NOT c.registered THEN c.total_amount ELSE 0 END), 0) as unregistered_total
		FROM members m
		LEFT JOIN contributions c ON c.member_id = m.id AND EXTRACT(YEAR FROM c.period) = $1
		WHERE m.scheme_id = $2
		GROUP BY m.id, m.member_no, m.first_name, m.last_name
		ORDER BY registered_total DESC
	`
	rows, err := s.db.QueryContext(ctx, query, year, schemeID)
	if err != nil {
		return nil, fmt.Errorf("query registered vs unregistered: %w", err)
	}
	defer rows.Close()

	var results []RegisteredVsUnregistered
	for rows.Next() {
		var r RegisteredVsUnregistered
		r.Year = year
		if err := rows.Scan(&r.MemberID, &r.MemberNo, &r.FullName, &r.RegisteredTotal, &r.UnregisteredTotal); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// ContributionTrend holds monthly trend data
type ContributionTrend struct {
	Month          string `json:"month"`
	EmployeeAmount int64  `json:"employee_amount"`
	EmployerAmount int64  `json:"employer_amount"`
	AVCAmount      int64  `json:"avc_amount"`
	TotalAmount    int64  `json:"total_amount"`
	MemberCount    int    `json:"member_count"`
}

// GetContributionTrends returns monthly contribution trends
func (s *Service) GetContributionTrends(ctx context.Context, schemeID string, year int) ([]ContributionTrend, error) {
	query := `
		SELECT TO_CHAR(period, 'YYYY-MM') as month,
		       COALESCE(SUM(employee_amount), 0) as employee_amount,
		       COALESCE(SUM(employer_amount), 0) as employer_amount,
		       COALESCE(SUM(avc_amount), 0) as avc_amount,
		       COALESCE(SUM(total_amount), 0) as total_amount,
		       COUNT(DISTINCT member_id) as member_count
		FROM contributions
		WHERE scheme_id = $1 AND EXTRACT(YEAR FROM period) = $2
		GROUP BY TO_CHAR(period, 'YYYY-MM')
		ORDER BY month
	`
	rows, err := s.db.QueryContext(ctx, query, schemeID, year)
	if err != nil {
		return nil, fmt.Errorf("query trends: %w", err)
	}
	defer rows.Close()

	var trends []ContributionTrend
	for rows.Next() {
		var t ContributionTrend
		if err := rows.Scan(&t.Month, &t.EmployeeAmount, &t.EmployerAmount, &t.AVCAmount, &t.TotalAmount, &t.MemberCount); err != nil {
			return nil, fmt.Errorf("scan trend: %w", err)
		}
		trends = append(trends, t)
	}
	return trends, rows.Err()
}

// AVCSummary holds AVC contribution summary
type AVCSummary struct {
	MemberID      string `json:"member_id"`
	MemberNo      string `json:"member_no"`
	FullName      string `json:"full_name"`
	AVCYearToDate int64  `json:"avc_year_to_date"`
	AVCCumulative int64  `json:"avc_cumulative"`
}

// GetAVCSummary returns AVC contribution summary
func (s *Service) GetAVCSummary(ctx context.Context, schemeID string, year int) ([]AVCSummary, error) {
	query := `
		SELECT m.id, m.member_no, m.first_name || ' ' || m.last_name,
		       COALESCE(SUM(CASE WHEN EXTRACT(YEAR FROM c.period) = $1 THEN c.avc_amount ELSE 0 END), 0) as avc_ytd,
		       COALESCE(SUM(c.avc_amount), 0) as avc_cumulative
		FROM members m
		LEFT JOIN contributions c ON c.member_id = m.id AND c.avc_amount > 0
		WHERE m.scheme_id = $2
		GROUP BY m.id, m.member_no, m.first_name, m.last_name
		HAVING COALESCE(SUM(c.avc_amount), 0) > 0
		ORDER BY avc_cumulative DESC
	`
	rows, err := s.db.QueryContext(ctx, query, year, schemeID)
	if err != nil {
		return nil, fmt.Errorf("query AVC summary: %w", err)
	}
	defer rows.Close()

	var summaries []AVCSummary
	for rows.Next() {
		var a AVCSummary
		if err := rows.Scan(&a.MemberID, &a.MemberNo, &a.FullName, &a.AVCYearToDate, &a.AVCCumulative); err != nil {
			return nil, fmt.Errorf("scan AVC: %w", err)
		}
		summaries = append(summaries, a)
	}
	return summaries, rows.Err()
}
