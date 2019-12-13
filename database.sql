CREATE TABLE IF NOT EXISTS games (
    id         INT(11) NOT NULL AUTO_INCREMENT,
    round      INT(11),
    prize      VARCHAR(50),
    used       TINYINT(1),
    status     VARCHAR(16),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    PRIMARY KEY(id)
)charset=utf8mb4;

CREATE TABLE IF NOT EXISTS messages (
    id                 INT(11) NOT NULL AUTO_INCREMENT,
    user_id            CHAR(36),
    conversation_id    CHAR(36),
    recipient_id       CHAR(36),
    message_id         CHAR(36),
    category           VARCHAR(32),
    data               LONGTEXT,     
    representative_id  CHAR(36),
    quote_message_id   CHAR(36),
    created_at         TIMESTAMP,
    PRIMARY KEY(id),
    UNIQUE  KEY(message_id)
)charset=utf8mb4;

CREATE TABLE IF NOT EXISTS payments (
    id                 INT(11) NOT NULL AUTO_INCREMENT,
    user_id            CHAR(36),
    asset_id           CHAR(36),
    trace_id           CHAR(36),
    amount             CHAR(50),
    paid               TINYINT(1),
    created_at         TIMESTAMP,
    updated_at         TIMESTAMP,
    PRIMARY KEY (id)
)charset=utf8mb4;

CREATE TABLE IF NOT EXISTS players (
    game_id            INT(11),
    user_id            CHAR(36),
    side               VARCHAR(36),
    conversation       CHAR(36),
    created_at         TIMESTAMP,
    updated_at         TIMESTAMP,
    PRIMARY KEY (game_id, user_id)
)charset=utf8mb4;

CREATE TABLE IF NOT EXISTS properties (
    `key`            VARCHAR(50),
    value            VARCHAR(255),
    updated_at       TIMESTAMP,
    PRIMARY KEY (`key`)
)charset=utf8mb4;

CREATE TABLE IF NOT EXISTS records(
    id         INT(11) NOT NULL AUTO_INCREMENT,
    game_id    INT(11),
    round      INT(11),
    one_red    INT(11),
    one_black  INT(11),
    two_red    INT(11),
    two_black  INT(11),
    created_at TIMESTAMP,
    PRIMARY KEY(id)
) charset=utf8mb4;

CREATE TABLE IF NOT EXISTS snapshots(
    snapshot_id     CHAR(36),
    amount          VARCHAR(50),
    trace_id        CHAR(36),
    user_id         CHAR(36),
    opponent_id     CHAR(36),
    data            CHAR(140),
    asset_id        CHAR(36),
    symbol          VARCHAR(16),  
    created_at      TIMESTAMP,
    PRIMARY KEY(snapshot_id)
) charset=utf8mb4;

CREATE TABLE IF NOT EXISTS transfers(
    transfer_id         CHAR(36),
    asset_id            CHAR(36),
    amount              VARCHAR(50),
    opponent_id         CHAR(36),
    memo                VARCHAR(140),
    created_at TIMESTAMP,
    PRIMARY KEY(transfer_id)
) charset=utf8mb4;

CREATE TABLE IF NOT EXISTS users(
    user_id            CHAR(36),
    paid_asset         CHAR(36),
    paid_amount        VARCHAR(50),
    earned_amount      VARCHAR(50),
    status             VARCHAR(16),
    created_at         TIMESTAMP,
    updated_at         TIMESTAMP,
    PRIMARY KEY(user_id)
) charset=utf8mb4;

