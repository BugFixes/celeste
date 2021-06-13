CREATE TABLE IF NOT EXISTS permission_group (
    id SERIAL,
    identifier VARCHAR(100),
    PRIMARY KEY(id)
);
INSERT INTO permission_group (
    identifier
) VALUES ('deity'),
         ('owner'),
         ('developer');
CREATE TABLE IF NOT EXISTS permission (
    id SERIAL,
    `key` VARCHAR(100),
    `action` VARCHAR(100),
    permission_group INT NOT NULL,
    CONSTRAINT fk_permission_group_id FOREIGN KEY(permission_group) REFERENCES permission_group(id)
);
INSERT INTO permission (
    `key`,
    `action`,
    permission_group
) VALUES ('developer', '*', 1),
         ('developer', 'create', 2),
         ('developer', 'delete', 2);

CREATE TABLE IF NOT EXISTS account (
    id SERIAL,
    identifier VARCHAR(100),
    account_key VARCHAR(100) NOT NULL,
    parent_id INT NULL,
    account_group INT NULL,
    PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS account_permission (
    id SERIAL,
    account_id INT NOT NULL,
    `key` VARCHAR(100),
    `action` VARCHAR(100),
    PRIMARY KEY(id),
    CONSTRAINT fk_account_id FOREIGN KEY(account_id) REFERENCES account(id)
);
