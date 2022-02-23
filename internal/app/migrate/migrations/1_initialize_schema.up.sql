CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

/********************
*  Nitro Type Logs  *
********************/

CREATE TABLE nt_api_team_logs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
	hash BYTEA NOT NULL UNIQUE,
	log_data JSON NOT NULL,

	deleted_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE nt_api_team_log_requests (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
	prev_id UUID REFERENCES nt_api_team_log_requests (id),
	api_team_log_id UUID NOT NULL REFERENCES nt_api_team_logs (id),
	response_type TEXT NOT NULL CHECK (response_type IN ('ERROR', 'CACHE', 'NEW')),	
	description TEXT NOT NULL,

	deleted_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

/********************
*  User Stats Data  *
********************/

CREATE TABLE users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
	reference_id INT NOT NULL UNIQUE,
	username TEXT NOT NULL UNIQUE,
	display_name TEXT NOT NULL,
	membership_type TEXT NOT NULL CHECK (membership_type IN ('BASIC', 'GOLD')),
	status TEXT NOT NULL CHECK (status IN ('NEW', 'ACTIVE', 'LEFT')),

	deleted_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_records (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
	request_id UUID NOT NULL REFERENCES nt_api_team_log_requests (id),
	user_id UUID NOT NULL REFERENCES users (id),

	played INT NOT NULL,
	typed INT NOT NULL,
	errs INT NOT NULL,
	secs INT NOT NULL,
	from_at TIMESTAMPTZ NOT NULL,
	to_at TIMESTAMPTZ NOT NULL,

	deleted_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

	UNIQUE (request_id, user_id)
);

/****************
*  Competition  *
****************/

CREATE TABLE competitions (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
	name TEXT NOT NULL,
	start_at TIMESTAMPTZ NOT NULL,
	finish_at TIMESTAMPTZ NOT NULL,

	deleted_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE competitions_to_user_records (
	competition_id UUID NOT NULL, 
	user_record_id UUID NOT NULL, 
	
	PRIMARY KEY (competition_id, user_record_id)
);

/************
*  Raffles  *
************/

CREATE TABLE raffles (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
	name TEXT UNIQUE,
	start_at TIMESTAMPTZ NOT NULL,
	finish_at TIMESTAMPTZ NOT NULL,

	deleted_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE raffle_tickets (
	code TEXT PRIMARY KEY
);

CREATE TABLE raffle_ticket_users (
	raffle_id UUID NOT NULL REFERENCES raffles (id),
	user_id UUID NOT NULL REFERENCES users (id),
	code TEXT NOT NULL REFERENCES raffle_tickets (code),
	reason TEXT,

	deleted_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

	PRIMARY KEY (raffle_id, user_id, code)
);
