package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/enielson/launchpad/pkg/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type templateRepository struct {
	db *sqlx.DB
}

// NewChainTemplateRepository creates a new PostgreSQL chain template repository
func NewChainTemplateRepository(db *sqlx.DB) interfaces.ChainTemplateRepository {
	return &templateRepository{db: db}
}

// GetByID retrieves a chain template by ID
func (r *templateRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ChainTemplate, error) {
	query := `
		SELECT id, template_name, template_description, template_category, supported_language,
			   default_token_supply, documentation_url, example_chains, version,
			   is_active, created_at, updated_at
		FROM chain_templates
		WHERE id = $1`

	var template models.ChainTemplate
	var documentationURL sql.NullString
	var exampleChains pq.StringArray

	err := r.db.QueryRowxContext(ctx, query, id).Scan(
		&template.ID,
		&template.TemplateName,
		&template.TemplateDescription,
		&template.TemplateCategory,
		&template.SupportedLanguage,
		&template.DefaultTokenSupply,
		&documentationURL,
		&exampleChains,
		&template.Version,
		&template.IsActive,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Handle nullable fields
	template.DocumentationURL = database.StringPtr(documentationURL)
	template.ExampleChains = []string(exampleChains)

	return &template, nil
}

// List retrieves chain templates with filtering and pagination
func (r *templateRepository) List(ctx context.Context, filters interfaces.TemplateFilters, pagination interfaces.Pagination) ([]models.ChainTemplate, int, error) {
	whereClause, args := r.buildTemplateWhereClause(filters)

	// Count query
	countQuery := "SELECT COUNT(*) FROM chain_templates" + whereClause
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count templates: %w", err)
	}

	// Data query
	dataQuery := fmt.Sprintf(`
		SELECT id, template_name, template_description, template_category, supported_language,
			   default_token_supply, documentation_url, example_chains, version,
			   is_active, created_at, updated_at
		FROM chain_templates
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause,
		len(args)+1,
		len(args)+2,
	)

	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := r.db.QueryxContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query templates: %w", err)
	}
	defer rows.Close()

	var templates []models.ChainTemplate
	for rows.Next() {
		var template models.ChainTemplate
		var documentationURL sql.NullString
		var exampleChains pq.StringArray

		err := rows.Scan(
			&template.ID,
			&template.TemplateName,
			&template.TemplateDescription,
			&template.TemplateCategory,
			&template.SupportedLanguage,
			&template.DefaultTokenSupply,
			&documentationURL,
			&exampleChains,
			&template.Version,
			&template.IsActive,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan template: %w", err)
		}

		// Handle nullable fields
		template.DocumentationURL = database.StringPtr(documentationURL)
		template.ExampleChains = []string(exampleChains)

		templates = append(templates, template)
	}

	return templates, total, nil
}

// GetByCategory retrieves templates by category
func (r *templateRepository) GetByCategory(ctx context.Context, category string) ([]models.ChainTemplate, error) {
	filters := interfaces.TemplateFilters{Category: category}
	pagination := interfaces.Pagination{Page: 1, Limit: 100, Offset: 0} // Get all for category
	templates, _, err := r.List(ctx, filters, pagination)
	return templates, err
}

// GetActive retrieves all active templates
func (r *templateRepository) GetActive(ctx context.Context) ([]models.ChainTemplate, error) {
	isActive := true
	filters := interfaces.TemplateFilters{IsActive: &isActive}
	pagination := interfaces.Pagination{Page: 1, Limit: 100, Offset: 0} // Get all active
	templates, _, err := r.List(ctx, filters, pagination)
	return templates, err
}

// Helper methods
func (r *templateRepository) buildTemplateWhereClause(filters interfaces.TemplateFilters) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argCount := 0

	if filters.Category != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("template_category = $%d", argCount))
		args = append(args, filters.Category)
	}

	if filters.IsActive != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argCount))
		args = append(args, *filters.IsActive)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	return whereClause, args
}
