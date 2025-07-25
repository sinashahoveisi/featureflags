package entity

import (
	"time"
)

// AuditAction represents the type of action performed on a flag
type AuditAction string

const (
	ActionCreate         AuditAction = "create"
	ActionEnable         AuditAction = "enable"
	ActionDisable        AuditAction = "disable"
	ActionCascadeDisable AuditAction = "cascade_disable"
	ActionUpdate         AuditAction = "update"
	ActionDelete         AuditAction = "delete"
)

// AuditLog represents a record of an action taken on a flag
type AuditLog struct {
	ID        int64       `json:"id" db:"id"`
	FlagID    int64       `json:"flag_id" db:"flag_id"`
	Action    AuditAction `json:"action" db:"action"`
	Actor     string      `json:"actor" db:"actor"`
	Reason    string      `json:"reason" db:"reason"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
}

// NewAuditLog creates a new audit log entry
func NewAuditLog(flagID int64, action AuditAction, actor, reason string) *AuditLog {
	return &AuditLog{
		FlagID:    flagID,
		Action:    action,
		Actor:     actor,
		Reason:    reason,
		CreatedAt: time.Now(),
	}
}

// IsCascadeAction returns true if the action is a cascade disable
func (a *AuditLog) IsCascadeAction() bool {
	return a.Action == ActionCascadeDisable
} 