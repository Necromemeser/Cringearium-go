package postgres

import "context"

func (r *UserRepository) EnsureAdministrator(ctx context.Context, passwordHash string) error {
	_, err := r.db.conn.ExecContext(ctx, `INSERT INTO users(username,email,password_hash,role)
 VALUES('admin','admin@mail.ru',$1,'admin')
 ON CONFLICT(email) DO UPDATE SET password_hash=EXCLUDED.password_hash,role='admin',updated_at=now()`, passwordHash)
	return err
}
