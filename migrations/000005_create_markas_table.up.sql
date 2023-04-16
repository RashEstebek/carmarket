CREATE TABLE IF NOT EXISTS markas (
                                      id bigserial PRIMARY KEY,
                                      name text NOT NULL,
                                      producer text NOT NULL,
                                      logo text NOT NULL
);