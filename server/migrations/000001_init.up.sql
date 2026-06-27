CREATE TABLE
  IF NOT EXISTS users (
    username VARCHAR(255) PRIMARY KEY,
    balance INT NOT NULL
  ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE
  IF NOT EXISTS polls (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    choice1 VARCHAR(255) NOT NULL,
    choice2 VARCHAR(255) NOT NULL,
    result INT NULL,
    due DATETIME NULL,
    created_by VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL
  ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE
  IF NOT EXISTS votes (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    poll_id BIGINT NOT NULL,
    username VARCHAR(255) NOT NULL,
    choice INT NOT NULL,
    bet INT NOT NULL,
    created_at DATETIME NOT NULL,
    UNIQUE KEY unique_votes_poll_id_username (poll_id, username),
    CONSTRAINT fk_votes_poll_id FOREIGN KEY (poll_id) REFERENCES polls (id)
  ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
