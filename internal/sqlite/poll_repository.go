package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"meetkat/internal/poll"
)

// PollRepository implements poll.Repository backed by SQLite.
type PollRepository struct {
	db *sql.DB
}

func NewPollRepository(db *sql.DB) *PollRepository {
	return &PollRepository{db: db}
}

func (r *PollRepository) Create(p *poll.Poll) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	answerMode := p.AnswerMode
	if answerMode == "" {
		answerMode = "yn"
	}
	res, err := tx.Exec(
		"INSERT INTO polls (public_id, admin_id, title, description, created_at, answer_mode) VALUES (?, ?, ?, ?, ?, ?)",
		p.ID, p.AdminID, p.Title, p.Description, p.CreatedAt.UTC().Format(time.RFC3339), answerMode,
	)
	if err != nil {
		return fmt.Errorf("insert poll: %w", err)
	}

	pollRowID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}

	for i, label := range p.Options {
		_, err := tx.Exec(
			"INSERT INTO poll_options (poll_id, label, position) VALUES (?, ?, ?)",
			pollRowID, label, i,
		)
		if err != nil {
			return fmt.Errorf("insert option %q: %w", label, err)
		}
	}

	return tx.Commit()
}

func (r *PollRepository) GetByPublicID(publicID string) (*poll.Poll, error) {
	return r.getPollByQuery(
		"SELECT id, public_id, admin_id, title, description, created_at, answer_mode FROM polls WHERE public_id = ?",
		publicID,
	)
}

func (r *PollRepository) GetByAdminID(adminID string) (*poll.Poll, error) {
	return r.getPollByQuery(
		"SELECT id, public_id, admin_id, title, description, created_at, answer_mode FROM polls WHERE admin_id = ?",
		adminID,
	)
}

func (r *PollRepository) getPollByQuery(query, value string) (*poll.Poll, error) {
	var rowID int64
	var p poll.Poll
	var createdAt string

	err := r.db.QueryRow(query, value).Scan(&rowID, &p.ID, &p.AdminID, &p.Title, &p.Description, &createdAt, &p.AnswerMode)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query poll: %w", err)
	}

	p.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

	// Load options ordered by position.
	optRows, err := r.db.Query(
		"SELECT id, label FROM poll_options WHERE poll_id = ? ORDER BY position",
		rowID,
	)
	if err != nil {
		return nil, fmt.Errorf("query options: %w", err)
	}
	defer func() { _ = optRows.Close() }()

	type optionRow struct {
		id    int64
		label string
	}
	var options []optionRow
	for optRows.Next() {
		var o optionRow
		if err := optRows.Scan(&o.id, &o.label); err != nil {
			return nil, fmt.Errorf("scan option: %w", err)
		}
		options = append(options, o)
		p.Options = append(p.Options, o.label)
	}
	if err := optRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate options: %w", err)
	}

	// Build a map from option_id -> label for vote responses.
	optionByID := make(map[int64]string, len(options))
	for _, o := range options {
		optionByID[o.id] = o.label
	}

	// Load votes.
	voteRows, err := r.db.Query(
		"SELECT id, name FROM votes WHERE poll_id = ? ORDER BY id",
		rowID,
	)
	if err != nil {
		return nil, fmt.Errorf("query votes: %w", err)
	}
	defer func() { _ = voteRows.Close() }()

	type voteRef struct {
		id   int64
		name string
	}
	var voteRefs []voteRef
	for voteRows.Next() {
		var v voteRef
		if err := voteRows.Scan(&v.id, &v.name); err != nil {
			return nil, fmt.Errorf("scan vote: %w", err)
		}
		voteRefs = append(voteRefs, v)
	}
	if err := voteRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate votes: %w", err)
	}

	// Load responses for each vote.
	for _, vr := range voteRefs {
		respRows, err := r.db.Query(
			"SELECT option_id, available FROM vote_responses WHERE vote_id = ?",
			vr.id,
		)
		if err != nil {
			return nil, fmt.Errorf("query responses for vote %d: %w", vr.id, err)
		}

		responses := make(map[string]string, len(options))
		for respRows.Next() {
			var optID int64
			var available int
			if err := respRows.Scan(&optID, &available); err != nil {
				_ = respRows.Close()
				return nil, fmt.Errorf("scan response: %w", err)
			}
			if label, ok := optionByID[optID]; ok {
				responses[label] = availableIntToString(available)
			}
		}
		_ = respRows.Close()
		if err := respRows.Err(); err != nil {
			return nil, fmt.Errorf("iterate responses: %w", err)
		}

		p.Votes = append(p.Votes, poll.Vote{Name: vr.name, Responses: responses})
	}

	return &p, nil
}

func (r *PollRepository) RemoveVote(pollID string, voterName string) error {
	res, err := r.db.Exec(
		"DELETE FROM votes WHERE poll_id = (SELECT id FROM polls WHERE public_id = ?) AND name = ?",
		pollID, voterName,
	)
	if err != nil {
		return fmt.Errorf("delete vote: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("vote not found")
	}
	return nil
}

func (r *PollRepository) AddVote(pollID string, vote poll.Vote) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Get the internal poll row ID.
	var rowID int64
	err = tx.QueryRow("SELECT id FROM polls WHERE public_id = ?", pollID).Scan(&rowID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("poll not found")
	}
	if err != nil {
		return fmt.Errorf("query poll id: %w", err)
	}

	// Insert the vote.
	res, err := tx.Exec("INSERT INTO votes (poll_id, name) VALUES (?, ?)", rowID, vote.Name)
	if err != nil {
		return fmt.Errorf("insert vote: %w", err)
	}
	voteID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}

	// Load option IDs for this poll, keyed by label.
	optRows, err := tx.Query("SELECT id, label FROM poll_options WHERE poll_id = ?", rowID)
	if err != nil {
		return fmt.Errorf("query options: %w", err)
	}
	defer func() { _ = optRows.Close() }()

	optionIDByLabel := make(map[string]int64)
	for optRows.Next() {
		var id int64
		var label string
		if err := optRows.Scan(&id, &label); err != nil {
			return fmt.Errorf("scan option: %w", err)
		}
		optionIDByLabel[label] = id
	}
	if err := optRows.Err(); err != nil {
		return fmt.Errorf("iterate options: %w", err)
	}

	// Insert vote responses.
	for label, value := range vote.Responses {
		optID, ok := optionIDByLabel[label]
		if !ok {
			continue
		}
		_, err := tx.Exec(
			"INSERT INTO vote_responses (vote_id, option_id, available) VALUES (?, ?, ?)",
			voteID, optID, availableStringToInt(value),
		)
		if err != nil {
			return fmt.Errorf("insert response for %q: %w", label, err)
		}
	}

	return tx.Commit()
}

func (r *PollRepository) Delete(pollID string) error {
	res, err := r.db.Exec("DELETE FROM polls WHERE public_id = ?", pollID)
	if err != nil {
		return fmt.Errorf("delete poll: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("poll not found")
	}
	return nil
}

func (r *PollRepository) UpdateVote(pollID string, oldName string, vote poll.Vote) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Get the internal poll row ID.
	var rowID int64
	err = tx.QueryRow("SELECT id FROM polls WHERE public_id = ?", pollID).Scan(&rowID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("poll not found")
	}
	if err != nil {
		return fmt.Errorf("query poll id: %w", err)
	}

	// Find the existing vote row ID (preserves id and voted_at).
	var voteID int64
	err = tx.QueryRow("SELECT id FROM votes WHERE poll_id = ? AND name = ?", rowID, oldName).Scan(&voteID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("vote not found")
	}
	if err != nil {
		return fmt.Errorf("query vote id: %w", err)
	}

	// Update the voter name and mark as edited, preserving id and voted_at.
	_, err = tx.Exec("UPDATE votes SET name = ?, edited_at = datetime('now') WHERE id = ?", vote.Name, voteID)
	if err != nil {
		return fmt.Errorf("update vote: %w", err)
	}

	// Load option IDs for this poll, keyed by label.
	optRows, err := tx.Query("SELECT id, label FROM poll_options WHERE poll_id = ?", rowID)
	if err != nil {
		return fmt.Errorf("query options: %w", err)
	}
	defer func() { _ = optRows.Close() }()

	optionIDByLabel := make(map[string]int64)
	for optRows.Next() {
		var id int64
		var label string
		if err := optRows.Scan(&id, &label); err != nil {
			return fmt.Errorf("scan option: %w", err)
		}
		optionIDByLabel[label] = id
	}
	if err := optRows.Err(); err != nil {
		return fmt.Errorf("iterate options: %w", err)
	}

	// Upsert vote responses.
	for label, value := range vote.Responses {
		optID, ok := optionIDByLabel[label]
		if !ok {
			continue
		}
		_, err := tx.Exec(
			"INSERT INTO vote_responses (vote_id, option_id, available) VALUES (?, ?, ?) ON CONFLICT(vote_id, option_id) DO UPDATE SET available = excluded.available",
			voteID, optID, availableStringToInt(value),
		)
		if err != nil {
			return fmt.Errorf("upsert response for %q: %w", label, err)
		}
	}

	return tx.Commit()
}

// availableStringToInt maps response strings to DB integers: "yes"->1, "maybe"->2, else->0.
func availableStringToInt(s string) int {
	switch s {
	case "yes":
		return 1
	case "maybe":
		return 2
	default:
		return 0
	}
}

// availableIntToString maps DB integers to response strings: 1->"yes", 2->"maybe", else->"no".
func availableIntToString(i int) string {
	switch i {
	case 1:
		return "yes"
	case 2:
		return "maybe"
	default:
		return "no"
	}
}
