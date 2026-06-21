-- Test schema mirroring the production tables (see project schema / sqlite
-- .schema). Kept here so tests stay hermetic and don't depend on a checked-in
-- database file.

CREATE TABLE categories(
	id integer primary key autoincrement,
	priority INTEGER not null,
	label text,
	is_ignored BOOLEAN DEFAULT 0,
	type TEXT CHECK(type IS NULL OR type IN ('income', 'fixed', 'fun', 'neutral'))
);

CREATE TABLE category_values(
	id integer primary key autoincrement,
	category_id integer not null,
	value text not null
);

CREATE TRIGGER validate_insert_categories
BEFORE INSERT ON categories
FOR EACH ROW
WHEN EXISTS (SELECT 1 FROM categories WHERE priority = NEW.priority)
BEGIN
	SELECT RAISE(ABORT, 'Error: This value already exists in the table.');
END;

CREATE TRIGGER validate_update_category_priority
BEFORE UPDATE OF priority ON categories
FOR EACH ROW
WHEN EXISTS (SELECT 1 FROM categories
			 WHERE priority = NEW.priority
			 AND id != OLD.id)
BEGIN
	SELECT RAISE(ABORT, 'Error: This value already exists in another row.');
END;

CREATE TABLE transactions(
	name text,
	amount int,
	date text,
	source text,
	account text,
	category text,
	id text PRIMARY KEY,
	description text,
	category_id integer,
	is_reimbursement BOOLEAN DEFAULT 0
);

CREATE TABLE net_worth(
	id text PRIMARY KEY,
	date text,
	cash real,
	investment real,
	debit real,
	credit real,
	savings real,
	retirement real,
	loans real
);

CREATE TABLE trades(
	id integer primary key autoincrement,
	ticker text not null,
	purchase_date text not null,
	shares integer not null,
	price integer not null,
	type text not null,
	account text not null,
	name text
);

CREATE TABLE kv_cache(
	key text primary key,
	value text,
	expires_at text
);
