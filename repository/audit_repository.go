package repository

import (
	"context"
	"fmt"

	"featureflags/entity"

	"github.com/jmoiron/sqlx"
)

type AuditRepository interface {
	CreateAuditLog(ctx context.Context, log *entity.AuditLog) error
	ListAuditLogsByFlagID(ctx context.Context, flagID int64) ([]*entity.AuditLog, error)
	ListAllAuditLogs(ctx context.Context, limit, offset int) ([]*entity.AuditLog, error)
}

type pgAuditRepository struct {
	db *sqlx.DB
}

func NewAuditRepository(db *sqlx.DB) AuditRepository {
	return &pgAuditRepository{db: db}
}

func (r *pgAuditRepository) CreateAuditLog(ctx context.Context, log *entity.AuditLog) error {
	query := `INSERT INTO audit_logs (flag_id, action, actor, reason) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, log.FlagID, log.Action, log.Actor, log.Reason)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

func (r *pgAuditRepository) ListAuditLogsByFlagID(ctx context.Context, flagID int64) ([]*entity.AuditLog, error) {
	var logs []*entity.AuditLog
	query := `
		SELECT id, flag_id, action, actor, reason, created_at 
		FROM audit_logs 
		WHERE flag_id = $1 
		ORDER BY created_at DESC
	`
	err := r.db.SelectContext(ctx, &logs, query, flagID)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs by flag ID: %w", err)
	}
	return logs, nil
}

func (r *pgAuditRepository) ListAllAuditLogs(ctx context.Context, limit, offset int) ([]*entity.AuditLog, error) {
	var logs []*entity.AuditLog
	query := `
		SELECT al.id, al.flag_id, al.action, al.actor, al.reason, al.created_at
		FROM audit_logs al
		ORDER BY al.created_at DESC
		LIMIT $1 OFFSET $2
	`
	err := r.db.SelectContext(ctx, &logs, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list all audit logs: %w", err)
	}
	return logs, nil
} 