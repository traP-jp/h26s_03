CREATE TABLE
  IF NOT EXISTS members (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL
  ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE
  IF NOT EXISTS projects (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    owner_member_id BIGINT NOT NULL,
    FOREIGN KEY (owner_member_id) REFERENCES members (id)
  ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE
  IF NOT EXISTS tasks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    project_id BIGINT NOT NULL,
    assignee_member_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    status VARCHAR(64) NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects (id),
    FOREIGN KEY (assignee_member_id) REFERENCES members (id)
  ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
