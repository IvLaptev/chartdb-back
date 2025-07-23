create type user_type as enum (
    'GUEST',
    'STUDENT',
    'TEACHER',
    'ADMIN'
);

create table users (
    id text primary key,
    login text not null,
    password_hash text,
    type user_type not null,
    confirmed_at timestamp with time zone,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
    deleted_at timestamp with time zone
);

insert into users (id, login, password_hash, type, created_at, updated_at)
select distinct user_id as user_id, user_id, null, 'GUEST'::user_type, now(), now() 
from diagrams;

alter table diagrams add constraint fk_diagrams_user_id foreign key (user_id) references users (id);

create table user_confirmations (
    id text primary key,
    user_id text not null,
    created_at timestamp with time zone not null,
    expires_at timestamp with time zone not null
);
