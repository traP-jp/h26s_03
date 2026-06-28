package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/traP-jp/h26s_03/server/internal/gen/openapi"
	"github.com/traP-jp/h26s_03/server/internal/middleware/authx"
)

const anonymousUser = "anonymous"

func (h *Handler) CreatePoll(ctx context.Context, req *openapi.CreatePollRequest) (*openapi.Poll, error) {
	createdBy, ok := authx.UserFromRequestContext(ctx)
	if !ok {
		createdBy = anonymousUser
	}

	createdAt := time.Now().UTC()
	due := sql.NullTime{}
	if v, ok := req.Due.Get(); ok {
		due = sql.NullTime{Time: v, Valid: true}
	}

	result, err := h.db.ExecContext(ctx, `
		INSERT INTO polls (name, choice1, choice2, result, due, created_by, created_at)
		VALUES (?, ?, ?, NULL, ?, ?, ?)
	`, req.Name, req.Choice1, req.Choice2, due, createdBy, createdAt)
	if err != nil {
		return nil, fmt.Errorf("create poll: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get created poll id: %w", err)
	}

	poll := &openapi.Poll{
		ID:        id,
		Name:      req.Name,
		Choice1:   req.Choice1,
		Choice2:   req.Choice2,
		Result:    nilInt(),
		Due:       nilDateTime(due),
		CreatedBy: createdBy,
		CreatedAt: createdAt,
	}

	return poll, nil
}

func (h *Handler) CreateVote(ctx context.Context, req *openapi.CreateVoteRequest, params openapi.CreateVoteParams) (openapi.CreateVoteRes, error) {
	username, ok := authx.UserFromRequestContext(ctx)
	if !ok {
		username = anonymousUser
	}

	if req.Choice != 1 && req.Choice != 2 {
		return &openapi.CreateVoteBadRequest{}, nil
	}

	tx, err := h.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create vote transaction: %w", err)
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRowxContext(ctx, `SELECT EXISTS(SELECT 1 FROM polls WHERE id = ?)`, params.ID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check poll exists: %w", err)
	}
	if exists == 0 {
		return &openapi.CreateVoteNotFound{}, nil
	}

	var alreadyVoted int
	if err := tx.QueryRowxContext(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM votes WHERE poll_id = ? AND username = ?)`,
		params.ID,
		username,
	).Scan(&alreadyVoted); err != nil {
		return nil, fmt.Errorf("check vote exists: %w", err)
	}
	if alreadyVoted != 0 {
		return &openapi.CreateVoteConflict{}, nil
	}

	var balance int
	if err := tx.QueryRowxContext(ctx, `SELECT balance FROM users WHERE username = ? FOR UPDATE`, username).Scan(&balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &openapi.CreateVoteConflict{}, nil
		}
		return nil, fmt.Errorf("get user balance: %w", err)
	}

	if balance < req.Bet {
		return &openapi.CreateVoteConflict{}, nil
	}

	createdAt := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `UPDATE users SET balance = balance - ? WHERE username = ?`, req.Bet, username); err != nil {
		return nil, fmt.Errorf("update user balance: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO votes (poll_id, username, choice, bet, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, params.ID, username, req.Choice, req.Bet, createdAt)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			switch mysqlErr.Number {
			case 1062:
				return &openapi.CreateVoteConflict{}, nil
			case 1452:
				return &openapi.CreateVoteNotFound{}, nil
			}
		}
		return nil, fmt.Errorf("create vote: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get created vote id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create vote transaction: %w", err)
	}

	return &openapi.Vote{
		ID:        id,
		Username:  username,
		Choice:    req.Choice,
		Bet:       req.Bet,
		CreatedAt: createdAt,
	}, nil
}

func (h *Handler) DeletePoll(ctx context.Context, params openapi.DeletePollParams) (openapi.DeletePollRes, error) {
	currentUser, ok := authx.UserFromRequestContext(ctx)
	if !ok {
		currentUser = anonymousUser
	}

	var createdBy string
	if err := h.db.QueryRowxContext(ctx, `SELECT created_by FROM polls WHERE id = ?`, params.ID).Scan(&createdBy); err != nil {
		if err == sql.ErrNoRows {
			return &openapi.DeletePollNotFound{}, nil
		}
		return nil, fmt.Errorf("get poll creator: %w", err)
	}
	if createdBy != currentUser {
		return &openapi.DeletePollForbidden{}, nil
	}

	tx, err := h.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin delete poll transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM votes WHERE poll_id = ?`, params.ID); err != nil {
		return nil, fmt.Errorf("delete poll votes: %w", err)
	}

	result, err := tx.ExecContext(ctx, `DELETE FROM polls WHERE id = ?`, params.ID)
	if err != nil {
		return nil, fmt.Errorf("delete poll: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get deleted poll count: %w", err)
	}
	if affected == 0 {
		return &openapi.DeletePollNotFound{}, nil
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit delete poll transaction: %w", err)
	}

	return &openapi.DeletePollNoContent{}, nil
}

func nilInt() openapi.NilInt {
	v := openapi.NilInt{}
	v.SetToNull()
	return v
}

func nilDateTime(t sql.NullTime) openapi.NilDateTime {
	v := openapi.NilDateTime{}
	if t.Valid {
		v.SetTo(t.Time)
		return v
	}
	v.SetToNull()
	return v
}

func (h *Handler) GetPoll(ctx context.Context, params openapi.GetPollParams) (openapi.GetPollRes, error) {
	var (
		poll      openapi.Poll
		result    sql.NullInt64
		due       sql.NullTime
		createdAt time.Time
	)

	if err := h.db.QueryRowxContext(
		ctx,
		"SELECT id, name, choice1, choice2, result, due, created_by, created_at FROM polls WHERE id = ?",
		params.ID,
	).Scan(
		&poll.ID,
		&poll.Name,
		&poll.Choice1,
		&poll.Choice2,
		&result,
		&due,
		&poll.CreatedBy,
		&createdAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return &openapi.GetPollNotFound{}, nil
		}
		return nil, err
	}

	if result.Valid {
		poll.Result.SetTo(int(result.Int64))
	} else {
		poll.Result.SetToNull()
	}

	if due.Valid {
		poll.Due.SetTo(due.Time)
	} else {
		poll.Due.SetToNull()
	}

	poll.CreatedAt = createdAt

	return &poll, nil
}

func (h *Handler) GetPollVotes(ctx context.Context, params openapi.GetPollVotesParams) (openapi.GetPollVotesRes, error) {
	var exists int
	if err := h.db.QueryRowxContext(ctx, `SELECT EXISTS(SELECT 1 FROM polls WHERE id = ?)`, params.ID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check poll exists: %w", err)
	}
	if exists == 0 {
		return &openapi.GetPollVotesNotFound{}, nil
	}

	rows, err := h.db.QueryxContext(ctx, `
		SELECT id, username, choice, bet, created_at
		FROM votes
		WHERE poll_id = ?
		ORDER BY id ASC
	`, params.ID)
	if err != nil {
		return nil, fmt.Errorf("get poll votes: %w", err)
	}
	defer rows.Close()

	votes := make([]openapi.Vote, 0)
	for rows.Next() {
		var vote openapi.Vote
		if err := rows.Scan(&vote.ID, &vote.Username, &vote.Choice, &vote.Bet, &vote.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan poll vote: %w", err)
		}
		votes = append(votes, vote)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate poll votes: %w", err)
	}

	return &openapi.VotesResponse{Data: votes}, nil
}
