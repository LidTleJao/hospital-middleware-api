-- Initial schema.
-- Tables are designed and written in the next step.
create table hospitals (
    hospital_id bigserial primary key,
    hospital_code varchar(50) not null ,
    hospital_name_th varchar(255) ,
    hospital_name_en varchar(255) ,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (hospital_code),
    unique (hospital_name_th),
    unique (hospital_name_en)
);

create table patients (
    patient_id bigserial primary key,
    hospital_id bigint not null references hospitals(hospital_id) on delete cascade,
    first_name_th varchar(255),
    middle_name_th varchar(255),
    last_name_th varchar(255),
    first_name_en varchar(255),
    middle_name_en varchar(255),
    last_name_en varchar(255),
    date_of_birth date not null,
    patient_hn varchar(255) not null ,
    national_id varchar(13) ,
    passport_id varchar(255) ,
    phone_number varchar(20),
    email varchar(255),
    gender varchar(10) not null ,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (hospital_id,patient_hn),
    unique (hospital_id,national_id),
    unique (hospital_id,passport_id),
    check (national_id ~ '^[0-9]{13}$'),
    check (passport_id is null or passport_id ~ '^[A-Z0-9]{5,20}$'),
    check (gender in ('M', 'F')),
    check (national_id is not null or passport_id is not null)
);

create index idx_patients_hospital_id_phone_number on patients(hospital_id, phone_number);
create index idx_patients_hospital_id_email on patients(hospital_id, email);
create index idx_patients_hospital_id_date_of_birth on patients(hospital_id, date_of_birth);
create index idx_patients_hospital_id_first_name_th on patients(hospital_id, first_name_th);
create index idx_patients_hospital_id_last_name_th on patients(hospital_id, last_name_th);
create index idx_patients_hospital_id_first_name_en on patients(hospital_id, first_name_en);
create index idx_patients_hospital_id_last_name_en on patients(hospital_id, last_name_en);


create table staffs (
    staff_id bigserial primary key,
    hospital_id bigint not null references hospitals(hospital_id) on delete cascade,
    username varchar(255) not null ,
    password varchar(255) not null, -- hashed password
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (hospital_id,username)
);

create or replace function update_updated_at_column()
returns trigger as $$
    begin
    new.updated_at = now();
    return new;
    end;
$$ language 'plpgsql';

create trigger update_hospitals_updated_at
    before update on hospitals
    for each row
    execute procedure update_updated_at_column();

create trigger update_patients_updated_at
    before update on patients
    for each row
    execute procedure update_updated_at_column();

create trigger update_staffs_updated_at
    before update on staffs
    for each row
    execute procedure update_updated_at_column();

insert into hospitals (hospital_code, hospital_name_th, hospital_name_en) values
('H001', 'โรงพยาบาล A', 'Hospital A'),
('H002', 'โรงพยาบาล B', 'Hospital B'),
('H003', 'โรงพยาบาล C', 'Hospital C'),
('H004', 'โรงพยาบาล D', 'Hospital D'),
('H005', 'โรงพยาบาล E', 'Hospital E');
