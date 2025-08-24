CREATE TABLE IF NOT EXISTS auth_challenges (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  public_key_hex TEXT NOT NULL,
  challenge TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  expires_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_auth_challenges_pub ON auth_challenges(public_key_hex);
