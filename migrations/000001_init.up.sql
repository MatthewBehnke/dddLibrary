CREATE TABLE books (
    id        uuid PRIMARY KEY,
    title     text NOT NULL,
    available boolean NOT NULL DEFAULT true
);

CREATE TABLE members (
    id           uuid PRIMARY KEY,
    name         text NOT NULL,
    status       text NOT NULL,
    active_loans integer NOT NULL DEFAULT 0 CHECK (active_loans >= 0)
);

CREATE TABLE loans (
    id          uuid PRIMARY KEY,
    book_id     uuid NOT NULL REFERENCES books (id),
    member_id   uuid NOT NULL REFERENCES members (id),
    borrowed_at timestamptz NOT NULL,
    due_date    timestamptz NOT NULL,
    returned_at timestamptz
);

CREATE INDEX loans_member_id_idx ON loans (member_id);
CREATE INDEX loans_book_id_idx ON loans (book_id);
