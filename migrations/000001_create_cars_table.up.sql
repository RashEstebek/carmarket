CREATE TABLE IF NOT EXISTS cars (
                                       id bigserial PRIMARY KEY,
                                       model text NOT NULL,
                                       year integer NOT NULL,
                                       price integer NOT NULL,
                                       marka text NOT NULL,
                                       color text NOT NULL,
                                       type text NOT NULL,
                                       image text NOT NULL,
                                       description text NOT NULL
);