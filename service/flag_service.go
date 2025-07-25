package service

import (
	"context"
	"errors"
	"fmt"

	"featureflags/entity"
	"featureflags/pkg/logger"
	"featureflags/repository"
	"featureflags/validator"
)

var (
	ErrMissingActiveDependencies = errors.New("missing active dependencies")
	ErrCircularDependency       = errors.New("circular dependency detected")
	ErrFlagNotFound            = errors.New("flag not found")
	ErrFlagAlreadyExists       = errors.New("flag already exists")
)

// DependencyError represents an error with missing dependencies
type DependencyError struct {
	Message             string   `json:"error"`
	MissingDependencies []string `json:"missing_dependencies"`
}

func (e DependencyError) Error() string {
	return e.Message
}

// FlagService defines the interface for flag business logic
type FlagService interface {
	CreateFlag(ctx context.Context, req validator.FlagCreateRequest, actor string) (*entity.Flag, error)
	EnableFlag(ctx context.Context, flagID int64, actor, reason string) error
	DisableFlag(ctx context.Context, flagID int64, actor, reason string) error
	ToggleFlag(ctx context.Context, flagID int64, req validator.FlagToggleRequest, actor string) error
	GetFlag(ctx context.Context, flagID int64) (*entity.Flag, error)
	ListFlags(ctx context.Context) ([]*entity.Flag, error)
	GetFlagAuditLogs(ctx context.Context, flagID int64) ([]*entity.AuditLog, error)
}

type flagService struct {
	flagRepo  repository.FlagRepository
	auditRepo repository.AuditRepository
	logger    *logger.Logger
}

func NewFlagService(flagRepo repository.FlagRepository, auditRepo repository.AuditRepository, log *logger.Logger) FlagService {
	return &flagService{
		flagRepo:  flagRepo,
		auditRepo: auditRepo,
		logger:    log,
	}
}

func (s *flagService) CreateFlag(ctx context.Context, req validator.FlagCreateRequest, actor string) (*entity.Flag, error) {
	// Validate request
	if err := validator.ValidateFlagCreateRequest(req); err != nil {
		s.logger.Warnw("Invalid flag creation request", "error", err, "actor", actor)
		return nil, err
	}

	// Validate actor
	if err := validator.ValidateActor(actor); err != nil {
		return nil, err
	}

	// Validate dependencies exist
	if len(req.Dependencies) > 0 {
		if err := s.validateDependenciesExist(ctx, req.Dependencies); err != nil {
			return nil, err
		}

		// Check for circular dependencies
		hasCircular, err := s.flagRepo.HasCircularDependency(ctx, 0, req.Dependencies)
		if err != nil {
			s.logger.Errorw("Failed to check circular dependency", "error", err)
			return nil, fmt.Errorf("failed to validate dependencies: %w", err)
		}
		if hasCircular {
			s.logger.Warnw("Circular dependency detected", "dependencies", req.Dependencies, "actor", actor)
			return nil, ErrCircularDependency
		}
	}

	// Create flag entity
	flag := &entity.Flag{
		Name:   req.Name,
		Status: entity.FlagDisabled, // Always start disabled
	}

	// Create flag in repository
	flagID, err := s.flagRepo.CreateFlag(ctx, flag)
	if err != nil {
		if errors.Is(err, repository.ErrFlagAlreadyExists) {
			return nil, ErrFlagAlreadyExists
		}
		s.logger.Errorw("Failed to create flag", "error", err, "name", req.Name)
		return nil, fmt.Errorf("failed to create flag: %w", err)
	}

	flag.ID = flagID

	// Add dependencies
	for _, depID := range req.Dependencies {
		if err := s.flagRepo.AddDependency(ctx, flagID, depID); err != nil {
			s.logger.Errorw("Failed to add dependency", "error", err, "flagID", flagID, "depID", depID)
			return nil, fmt.Errorf("failed to add dependency: %w", err)
		}
	}

	flag.Dependencies = req.Dependencies

	// Create audit log
	auditLog := entity.NewAuditLog(flagID, entity.ActionCreate, actor, "Flag created")
	if err := s.auditRepo.CreateAuditLog(ctx, auditLog); err != nil {
		s.logger.Warnw("Failed to create audit log", "error", err, "flagID", flagID)
	}

	s.logger.Infow("Flag created successfully", "flagID", flagID, "name", req.Name, "actor", actor)
	return flag, nil
}

func (s *flagService) EnableFlag(ctx context.Context, flagID int64, actor, reason string) error {
	if err := validator.ValidateFlagID(flagID); err != nil {
		return err
	}
	if err := validator.ValidateActor(actor); err != nil {
		return err
	}

	// Get flag with dependencies
	flag, err := s.flagRepo.GetFlagByID(ctx, flagID)
	if err != nil {
		if errors.Is(err, repository.ErrFlagNotFound) {
			return ErrFlagNotFound
		}
		return fmt.Errorf("failed to get flag: %w", err)
	}

	// Check if already enabled
	if flag.IsEnabled() {
		return nil // Already enabled, no-op
	}

	// Validate dependencies are enabled
	if flag.HasDependencies() {
		missingDeps, err := s.getMissingActiveDependencies(ctx, flag.Dependencies)
		if err != nil {
			return fmt.Errorf("failed to check dependencies: %w", err)
		}
		if len(missingDeps) > 0 {
			s.logger.Warnw("Cannot enable flag due to missing dependencies", 
				"flagID", flagID, "missingDeps", missingDeps, "actor", actor)
			return DependencyError{
				Message:             "Missing active dependencies",
				MissingDependencies: missingDeps,
			}
		}
	}

	// Enable flag
	if err := s.flagRepo.UpdateFlagStatus(ctx, flagID, entity.FlagEnabled); err != nil {
		s.logger.Errorw("Failed to enable flag", "error", err, "flagID", flagID)
		return fmt.Errorf("failed to enable flag: %w", err)
	}

	// Create audit log
	auditLog := entity.NewAuditLog(flagID, entity.ActionEnable, actor, reason)
	if err := s.auditRepo.CreateAuditLog(ctx, auditLog); err != nil {
		s.logger.Warnw("Failed to create audit log", "error", err, "flagID", flagID)
	}

	s.logger.Infow("Flag enabled successfully", "flagID", flagID, "actor", actor, "reason", reason)
	return nil
}

func (s *flagService) DisableFlag(ctx context.Context, flagID int64, actor, reason string) error {
	if err := validator.ValidateFlagID(flagID); err != nil {
		return err
	}
	if err := validator.ValidateActor(actor); err != nil {
		return err
	}

	// Get flag
	flag, err := s.flagRepo.GetFlagByID(ctx, flagID)
	if err != nil {
		if errors.Is(err, repository.ErrFlagNotFound) {
			return ErrFlagNotFound
		}
		return fmt.Errorf("failed to get flag: %w", err)
	}

	// Check if already disabled
	if flag.IsDisabled() {
		return nil // Already disabled, no-op
	}

	// Disable flag
	if err := s.flagRepo.UpdateFlagStatus(ctx, flagID, entity.FlagDisabled); err != nil {
		s.logger.Errorw("Failed to disable flag", "error", err, "flagID", flagID)
		return fmt.Errorf("failed to disable flag: %w", err)
	}

	// Create audit log
	auditLog := entity.NewAuditLog(flagID, entity.ActionDisable, actor, reason)
	if err := s.auditRepo.CreateAuditLog(ctx, auditLog); err != nil {
		s.logger.Warnw("Failed to create audit log", "error", err, "flagID", flagID)
	}

	// Cascade disable dependents
	if err := s.cascadeDisableDependents(ctx, flagID); err != nil {
		s.logger.Errorw("Failed to cascade disable dependents", "error", err, "flagID", flagID)
		// Don't return error, as the main flag was disabled successfully
	}

	s.logger.Infow("Flag disabled successfully", "flagID", flagID, "actor", actor, "reason", reason)
	return nil
}

func (s *flagService) ToggleFlag(ctx context.Context, flagID int64, req validator.FlagToggleRequest, actor string) error {
	if err := validator.ValidateFlagToggleRequest(req); err != nil {
		return err
	}

	if req.Enable {
		return s.EnableFlag(ctx, flagID, actor, req.Reason)
	}
	return s.DisableFlag(ctx, flagID, actor, req.Reason)
}

func (s *flagService) GetFlag(ctx context.Context, flagID int64) (*entity.Flag, error) {
	if err := validator.ValidateFlagID(flagID); err != nil {
		return nil, err
	}

	flag, err := s.flagRepo.GetFlagByID(ctx, flagID)
	if err != nil {
		if errors.Is(err, repository.ErrFlagNotFound) {
			return nil, ErrFlagNotFound
		}
		return nil, fmt.Errorf("failed to get flag: %w", err)
	}

	return flag, nil
}

func (s *flagService) ListFlags(ctx context.Context) ([]*entity.Flag, error) {
	flags, err := s.flagRepo.GetFlagsWithDependencies(ctx)
	if err != nil {
		s.logger.Errorw("Failed to list flags", "error", err)
		return nil, fmt.Errorf("failed to list flags: %w", err)
	}

	return flags, nil
}

func (s *flagService) GetFlagAuditLogs(ctx context.Context, flagID int64) ([]*entity.AuditLog, error) {
	if err := validator.ValidateFlagID(flagID); err != nil {
		return nil, err
	}

	// Verify flag exists
	_, err := s.flagRepo.GetFlagByID(ctx, flagID)
	if err != nil {
		if errors.Is(err, repository.ErrFlagNotFound) {
			return nil, ErrFlagNotFound
		}
		return nil, fmt.Errorf("failed to verify flag existence: %w", err)
	}

	logs, err := s.auditRepo.ListAuditLogsByFlagID(ctx, flagID)
	if err != nil {
		s.logger.Errorw("Failed to get audit logs", "error", err, "flagID", flagID)
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return logs, nil
}

// validateDependenciesExist checks if all dependency IDs exist
func (s *flagService) validateDependenciesExist(ctx context.Context, dependencyIDs []int64) error {
	for _, depID := range dependencyIDs {
		_, err := s.flagRepo.GetFlagByID(ctx, depID)
		if err != nil {
			if errors.Is(err, repository.ErrFlagNotFound) {
				return fmt.Errorf("dependency flag with ID %d not found", depID)
			}
			return fmt.Errorf("failed to validate dependency %d: %w", depID, err)
		}
	}
	return nil
}

// getMissingActiveDependencies returns the names of dependencies that are not enabled
func (s *flagService) getMissingActiveDependencies(ctx context.Context, dependencyIDs []int64) ([]string, error) {
	var missingDeps []string

	for _, depID := range dependencyIDs {
		flag, err := s.flagRepo.GetFlagByID(ctx, depID)
		if err != nil {
			return nil, fmt.Errorf("failed to get dependency flag %d: %w", depID, err)
		}
		if flag.IsDisabled() {
			missingDeps = append(missingDeps, flag.Name)
		}
	}

	return missingDeps, nil
}

// cascadeDisableDependents disables all flags that depend on this flag
func (s *flagService) cascadeDisableDependents(ctx context.Context, flagID int64) error {
	dependents, err := s.flagRepo.GetDependents(ctx, flagID)
	if err != nil {
		return fmt.Errorf("failed to get dependents: %w", err)
	}

	for _, depID := range dependents {
		// Get dependent flag to check if it's enabled
		depFlag, err := s.flagRepo.GetFlagByID(ctx, depID)
		if err != nil {
			s.logger.Errorw("Failed to get dependent flag", "error", err, "depID", depID)
			continue
		}

		if depFlag.IsEnabled() {
			// Disable the dependent flag
			if err := s.flagRepo.UpdateFlagStatus(ctx, depID, entity.FlagDisabled); err != nil {
				s.logger.Errorw("Failed to cascade disable dependent", "error", err, "depID", depID)
				continue
			}

			// Create audit log for cascade disable
			auditLog := entity.NewAuditLog(depID, entity.ActionCascadeDisable, "system", 
				fmt.Sprintf("Automatically disabled due to dependency flag %d being disabled", flagID))
			if err := s.auditRepo.CreateAuditLog(ctx, auditLog); err != nil {
				s.logger.Warnw("Failed to create cascade audit log", "error", err, "depID", depID)
			}

			s.logger.Infow("Cascade disabled dependent flag", "depID", depID, "parentFlagID", flagID)

			// Recursively disable dependents of this flag
			if err := s.cascadeDisableDependents(ctx, depID); err != nil {
				s.logger.Errorw("Failed to recursively cascade disable", "error", err, "depID", depID)
			}
		}
	}

	return nil
} 