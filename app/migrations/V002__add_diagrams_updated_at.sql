alter table diagrams add column updated_at timestamp with time zone;
update diagrams set updated_at = created_at;
alter table diagrams alter column updated_at set not null;
