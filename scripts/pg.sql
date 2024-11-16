CREATE TABLE "Tasks" (
    name VARCHAR(256) NOT NULL PRIMARY KEY,
    last_processed_block_number BIGINT DEFAULT 0,
    last_processed_block_timestamp BIGINT DEFAULT 0
);

CREATE TABLE "Logs" (
    chain_id BIGINT NOT NULL,
    block_number BIGINT NOT NULL,
    block_hash VARCHAR(256) NOT NULL,
    data TEXT NOT NULL,
    topics TEXT[] NOT NULL,
    txn_hash VARCHAR(256) NOT NULL,
    log_index BIGINT NOT NULL,
    removed BOOLEAN NOT NULL,
    timestamp BIGINT NOT NULL,
    PRIMARY KEY (chain_id, txn_hash, log_index)
);
