CREATE TABLE IF NOT EXISTS permissions (
    id bigserial PRIMARY KEY,
    code text NOT NULL
);

CREATE TABLE IF NOT EXISTS users_permissions (
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    permission_id bigint NOT NULL REFERENCES permissions ON DELETE CASCADE,
    -- This is a composite primary key
    -- It means that the combination of user_id and permission_id must be unique
    PRIMARY KEY (user_id, permission_id)
);

INSERT INTO permissions (code) VALUES ('movies:read'), ('movies:write');