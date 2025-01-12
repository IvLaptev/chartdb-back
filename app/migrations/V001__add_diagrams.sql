create extension if not exists btree_gist;

create table diagrams (
	id varchar(10) primary key,
	user_id text not null,
	client_diagram_id varchar(4) not null,
	code varchar(4) not null,
	object_storage_key varchar(20) not null,
	created_at timestamp with time zone not null,
	deleted_at timestamp with time zone,
	exclude using gist (code with =, user_id with <>) where (deleted_at is null)
);

create unique index idx_unique_object_storage_key
	on diagrams (object_storage_key) where (deleted_at is null);
