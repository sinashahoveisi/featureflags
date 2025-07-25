package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"featureflags/entity"

	"github.com/jmoiron/sqlx"
)

var (
	ErrFlagNotFound      = errors.New("flag not found")
	ErrFlagAlreadyExists = errors.New("flag already exists")
	ErrCircularDependency = errors.New("circular dependency detected")
)

// FlagRepository defines the interface for interacting with flag data
type FlagRepository interface {
	CreateFlag(ctx context.Context, flag *entity.Flag) (int64, error)
	GetFlagByID(ctx context.Context, id int64) (*entity.Flag, error)
	GetFlagByName(ctx context.Context, name string) (*entity.Flag, error)
	ListFlags(ctx context.Context) ([]*entity.Flag, error)
	UpdateFlagStatus(ctx context.Context, id int64, status entity.FlagStatus) error
	AddDependency(ctx context.Context, flagID, dependsOnID int64) error
	GetDependencies(ctx context.Context, flagID int64) ([]int64, error)
	GetDependents(ctx context.Context, flagID int64) ([]int64, error)
	HasCircularDependency(ctx context.Context, flagID int64, dependencyIDs []int64) (bool, error)
	GetFlagsWithDependencies(ctx context.Context) ([]*entity.Flag, error)
}

type pgFlagRepository struct {
	db *sqlx.DB
}

func NewFlagRepository(db *sqlx.DB) FlagRepository {
	return &pgFlagRepository{db: db}
}

func (r *pgFlagRepository) CreateFlag(ctx context.Context, flag *entity.Flag) (int64, error) {
	// Check if flag with same name already exists
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM flags WHERE name = $1", flag.Name)
	if err != nil {
		return 0, fmt.Errorf("failed to check flag existence: %w", err)
	}
	if count > 0 {
		return 0, ErrFlagAlreadyExists
	}

	query := `INSERT INTO flags (name, status) VALUES ($1, $2) RETURNING id`
	var flagID int64
	err = r.db.QueryRowContext(ctx, query, flag.Name, flag.Status).Scan(&flagID)
	if err != nil {
		return 0, fmt.Errorf("failed to create flag: %w", err)
	}
	return flagID, nil
}

func (r *pgFlagRepository) GetFlagByID(ctx context.Context, id int64) (*entity.Flag, error) {
	var flag entity.Flag
	query := `SELECT id, name, status, created_at, updated_at FROM flags WHERE id = $1`
	err := r.db.GetContext(ctx, &flag, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFlagNotFound
		}
		return nil, fmt.Errorf("failed to get flag by ID: %w", err)
	}
	
	// Load dependencies
	dependencies, err := r.GetDependencies(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load dependencies: %w", err)
	}
	flag.Dependencies = dependencies
	
	return &flag, nil
}

func (r *pgFlagRepository) GetFlagByName(ctx context.Context, name string) (*entity.Flag, error) {
	var flag entity.Flag
	query := `SELECT id, name, status, created_at, updated_at FROM flags WHERE name = $1`
	err := r.db.GetContext(ctx, &flag, query, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFlagNotFound
		}
		return nil, fmt.Errorf("failed to get flag by name: %w", err)
	}
	
	// Load dependencies
	dependencies, err := r.GetDependencies(ctx, flag.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load dependencies: %w", err)
	}
	flag.Dependencies = dependencies
	
	return &flag, nil
}

func (r *pgFlagRepository) ListFlags(ctx context.Context) ([]*entity.Flag, error) {
	var flags []*entity.Flag
	query := `SELECT id, name, status, created_at, updated_at FROM flags ORDER BY name`
	err := r.db.SelectContext(ctx, &flags, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list flags: %w", err)
	}
	return flags, nil
}

func (r *pgFlagRepository) GetFlagsWithDependencies(ctx context.Context) ([]*entity.Flag, error) {
	flags, err := r.ListFlags(ctx)
	if err != nil {
		return nil, err
	}
	
	// Load dependencies for each flag
	for _, flag := range flags {
		dependencies, err := r.GetDependencies(ctx, flag.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load dependencies for flag %d: %w", flag.ID, err)
		}
		flag.Dependencies = dependencies
	}
	
	return flags, nil
}

func (r *pgFlagRepository) UpdateFlagStatus(ctx context.Context, id int64, status entity.FlagStatus) error {
	query := `UPDATE flags SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update flag status: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrFlagNotFound
	}
	
	return nil
}

func (r *pgFlagRepository) AddDependency(ctx context.Context, flagID, dependsOnID int64) error {
	query := `INSERT INTO flag_dependencies (flag_id, depends_on_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, flagID, dependsOnID)
	if err != nil {
		return fmt.Errorf("failed to add dependency: %w", err)
	}
	return nil
}

func (r *pgFlagRepository) GetDependencies(ctx context.Context, flagID int64) ([]int64, error) {
	var dependencyIDs []int64
	query := `SELECT depends_on_id FROM flag_dependencies WHERE flag_id = $1 ORDER BY depends_on_id`
	err := r.db.SelectContext(ctx, &dependencyIDs, query, flagID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", err)
	}
	return dependencyIDs, nil
}

func (r *pgFlagRepository) GetDependents(ctx context.Context, flagID int64) ([]int64, error) {
	var dependentIDs []int64
	query := `SELECT flag_id FROM flag_dependencies WHERE depends_on_id = $1 ORDER BY flag_id`
	err := r.db.SelectContext(ctx, &dependentIDs, query, flagID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependents: %w", err)
	}
	return dependentIDs, nil
}

func (r *pgFlagRepository) HasCircularDependency(ctx context.Context, flagID int64, dependencyIDs []int64) (bool, error) {
	// For each proposed dependency, check if it would create a cycle
	for _, depID := range dependencyIDs {
		// Use recursive CTE to check if flagID is reachable from depID
		query := `
			WITH RECURSIVE dependency_path AS (
				-- Base case: direct dependencies of depID
				SELECT depends_on_id as id, 1 as depth
				FROM flag_dependencies 
				WHERE flag_id = $1
				
				UNION ALL
				
				-- Recursive case: follow the dependency chain
				SELECT fd.depends_on_id, dp.depth + 1
				FROM flag_dependencies fd
				JOIN dependency_path dp ON fd.flag_id = dp.id
				WHERE dp.depth < 10 -- Prevent infinite recursion
			)
			SELECT 1 FROM dependency_path WHERE id = $2 LIMIT 1
		`
		
		var exists int
		err := r.db.QueryRowContext(ctx, query, depID, flagID).Scan(&exists)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("failed to check circular dependency: %w", err)
		}
		if exists == 1 {
			return true, nil
		}
	}
	
	return false, nil
} 