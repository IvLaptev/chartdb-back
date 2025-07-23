alter table diagrams add column name text, add column tables_count bigint;
update diagrams set name = '', tables_count = 0;
alter table diagrams alter column name set not null, alter column tables_count set not null;
