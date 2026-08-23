DROP TABLE member_duplicate_audits;

DROP INDEX idx_members_normalized_phone;
DROP INDEX idx_members_normalized_name;

ALTER TABLE members
    DROP COLUMN normalized_congregation,
    DROP COLUMN normalized_city,
    DROP COLUMN normalized_neighborhood,
    DROP COLUMN normalized_address_number,
    DROP COLUMN normalized_address,
    DROP COLUMN normalized_phone,
    DROP COLUMN normalized_name;
