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
    key VARCHAR(100),
    action VARCHAR(100),
    permission_group INT NOT NULL,
    CONSTRAINT fk_permission_group_id FOREIGN KEY(permission_group) REFERENCES permission_group(id)
);
INSERT INTO permission (
    key,
    action,
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
    key VARCHAR(100),
    action VARCHAR(100),
    PRIMARY KEY(id),
    CONSTRAINT fk_account_id FOREIGN KEY(account_id) REFERENCES account(id)
);

CREATE TABLE IF NOT EXISTS frontend_versions (
    id SERIAL,
    version VARCHAR(100),
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS agent (
    id SERIAL,
    account_id INT NOT NULL,
    name VARCHAR(100),
    key UUID,
    secret UUID,
    PRIMARY KEY (id),
    CONSTRAINT fk_account_id FOREIGN KEY (account_id) REFERENCES account(id)
);

CREATE TABLE IF NOT EXISTS ticketing_details (
    id SERIAL,
    agent_id INT NOT NULL,
    system VARCHAR(100),
    details JSON,
    PRIMARY KEY (id),
    CONSTRAINT fk_agent_id FOREIGN KEY (agent_id) REFERENCES agent(id)
);

CREATE TABLE IF NOT EXISTS ticket (
    id SERIAL,
    agent_id INT NOT NULL,
    remote_id VARCHAR(100),
    system VARCHAR(100),
    hash TEXT,
    PRIMARY KEY (id),
    CONSTRAINT fk_agent_id FOREIGN KEY (agent_id) REFERENCES agent(id)
);
CREATE INDEX idx_tickets ON ticket(hash);


