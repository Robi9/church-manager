CREATE TABLE members (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    status TEXT,

    member_since TIMESTAMP,
    baptized BOOLEAN,
    baptism_date TIMESTAMP,

    church_role TEXT,
    marital_status TEXT,
    origin_denomination TEXT,

    membership_course_completed BOOLEAN,
    membership_course_completed_at TIMESTAMP,

    contacted BOOLEAN,
    contacted_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    created_by BIGINT
);