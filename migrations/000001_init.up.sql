-- Initial schema.
-- Tables are designed and written in the next step.
create table hospitals (
    hospital_id bigserial primary key,
    hospital_code varchar(50) not null ,
    hospital_name_th varchar(255) ,
    hospital_name_en varchar(255) ,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz,
    unique (hospital_code),
    unique (hospital_name_th),
    unique (hospital_name_en)
);

create table patients (
    patient_id bigserial primary key,
    hospital_id bigint not null references hospitals(hospital_id) on delete cascade,
    first_name_th varchar(255) not null,
    middle_name_th varchar(255),
    last_name_th varchar(255) not null,
    first_name_en varchar(255) not null,
    middle_name_en varchar(255),
    last_name_en varchar(255) not null,
    date_of_birth date not null,
    patient_hn varchar(255) not null ,
    national_id varchar(13) ,
    passport_id varchar(255) ,
    phone_number varchar(20),
    email varchar(255),
    gender varchar(10) not null ,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz,
    unique (hospital_id,patient_hn),
    unique (hospital_id,national_id),
    unique (hospital_id,passport_id),
    check (national_id ~ '^[0-9]{13}$'),
    check (passport_id is null or passport_id ~ '^[A-Z0-9]{5,20}$'),
    check (gender in ('M', 'F')),
    check (national_id is not null or passport_id is not null)
);

create table staffs (
    staff_id bigserial primary key,
    hospital_id bigint not null references hospitals(hospital_id) on delete cascade,
    username varchar(255) not null ,
    password varchar(255) not null, -- hashed password
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz,
    unique (hospital_id,username)
);



