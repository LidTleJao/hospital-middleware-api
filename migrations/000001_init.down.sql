-- Reverses 000001_init.up.sql.
drop trigger if exists update_hospitals_updated_at on hospitals;
drop trigger if exists update_patients_updated_at on patients;
drop trigger if exists update_staffs_updated_at on staffs;
drop function if exists update_updated_at_column();
drop table if exists staffs;
drop table if exists patients;
drop table if exists hospitals;
