ALTER TABLE members
    ADD COLUMN normalized_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN normalized_phone TEXT NOT NULL DEFAULT '',
    ADD COLUMN normalized_address TEXT NOT NULL DEFAULT '',
    ADD COLUMN normalized_address_number TEXT NOT NULL DEFAULT '',
    ADD COLUMN normalized_neighborhood TEXT NOT NULL DEFAULT '',
    ADD COLUMN normalized_city TEXT NOT NULL DEFAULT '',
    ADD COLUMN normalized_congregation TEXT NOT NULL DEFAULT '';

UPDATE members
SET
    normalized_name = translate(
        regexp_replace(lower(btrim(COALESCE(name, ''))), '[[:space:]]+', ' ', 'g'),
        'áàâãäåéèêëíìîïóòôõöúùûüçñýÿ',
        'aaaaaaeeeeiiiiooooouuuucnyy'
    ),
    normalized_phone = CASE
        WHEN regexp_replace(COALESCE(phone, ''), '[^0-9]', '', 'g') ~ '^55[0-9]{10,11}$'
            THEN substring(regexp_replace(COALESCE(phone, ''), '[^0-9]', '', 'g') FROM 3)
        ELSE regexp_replace(COALESCE(phone, ''), '[^0-9]', '', 'g')
    END,
    normalized_address = translate(
        regexp_replace(lower(btrim(COALESCE(address, ''))), '[[:space:]]+', ' ', 'g'),
        'áàâãäåéèêëíìîïóòôõöúùûüçñýÿ',
        'aaaaaaeeeeiiiiooooouuuucnyy'
    ),
    normalized_address_number = translate(
        regexp_replace(lower(btrim(COALESCE(address_number, ''))), '[[:space:]]+', ' ', 'g'),
        'áàâãäåéèêëíìîïóòôõöúùûüçñýÿ',
        'aaaaaaeeeeiiiiooooouuuucnyy'
    ),
    normalized_neighborhood = translate(
        regexp_replace(lower(btrim(COALESCE(neighborhood, ''))), '[[:space:]]+', ' ', 'g'),
        'áàâãäåéèêëíìîïóòôõöúùûüçñýÿ',
        'aaaaaaeeeeiiiiooooouuuucnyy'
    ),
    normalized_city = translate(
        regexp_replace(lower(btrim(COALESCE(city, ''))), '[[:space:]]+', ' ', 'g'),
        'áàâãäåéèêëíìîïóòôõöúùûüçñýÿ',
        'aaaaaaeeeeiiiiooooouuuucnyy'
    ),
    normalized_congregation = translate(
        regexp_replace(lower(btrim(COALESCE(congregation, ''))), '[[:space:]]+', ' ', 'g'),
        'áàâãäåéèêëíìîïóòôõöúùûüçñýÿ',
        'aaaaaaeeeeiiiiooooouuuucnyy'
    );

CREATE INDEX idx_members_normalized_name ON members(normalized_name)
    WHERE normalized_name <> '';
CREATE INDEX idx_members_normalized_phone ON members(normalized_phone)
    WHERE normalized_phone <> '';

CREATE TABLE member_duplicate_audits (
    id BIGSERIAL PRIMARY KEY,
    member_id BIGINT NOT NULL REFERENCES members(id),
    candidate_member_id BIGINT NOT NULL REFERENCES members(id),
    confirmed_by BIGINT NOT NULL REFERENCES users(id),
    score INTEGER NOT NULL,
    matched_fields TEXT[] NOT NULL,
    operation TEXT NOT NULL CHECK (operation IN ('create', 'update')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_member_duplicate_audits_member_id
    ON member_duplicate_audits(member_id);

