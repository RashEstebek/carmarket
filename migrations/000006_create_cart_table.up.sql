CREATE TABLE IF NOT EXISTS carts (
                                     user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
                                     cars_id bigint[],
                                     PRIMARY KEY (user_id)
    );