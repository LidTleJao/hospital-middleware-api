-- Demo patients, so a fresh `docker compose up` has something to search.
--
-- In normal operation these rows arrive from a hospital's HIS through
-- POST /patient/import. The HIS host named in the assignment does not
-- resolve, so this seed stands in for it and keeps the search endpoint
-- demonstrable on a clean checkout.
--
-- The rows cover the cases the schema was designed around:
--   A-001 / B-001  the same person, same national_id, at two hospitals
--   A-004          a foreign patient: no Thai name, no national_id
--   A-005          a Thai patient holding both a national id and a passport
--   A-002          middle names present in both languages
--   C-001          a hospital that has patients but no staff account

INSERT INTO patients (
    hospital_id,
    first_name_th, middle_name_th, last_name_th,
    first_name_en, middle_name_en, last_name_en,
    date_of_birth, patient_hn, national_id, passport_id,
    phone_number, email, gender
) VALUES
    (1, 'สมชาย', NULL,  'ใจดี',     'Somchai', NULL,      'Jaidee',    '1990-01-15', 'A-001', '1101700200111', NULL,        '0811111111', 'somchai@example.com',    'M'),
    (1, 'มานี',  'ศรี', 'รักไทย',   'Manee',   'Sri',     'Rakthai',   '1995-05-05', 'A-002', '1101700200222', NULL,        '0822222222', 'manee@example.com',      'F'),
    (1, 'สมหญิง', NULL, 'ดีงาม',    'Somying', NULL,      'Deengam',   '1988-11-30', 'A-003', '1101700200333', NULL,        '0833333333', NULL,                     'F'),
    (1, NULL,     NULL, NULL,        'John',    'Michael', 'Smith',     '1985-03-20', 'A-004', NULL,            'AB1234567', '0844444444', 'john.smith@example.com', 'M'),
    (1, 'ปิติ',   NULL, 'มั่งมี',    'Piti',    NULL,      'Mangmee',   '2000-07-07', 'A-005', '1101700200555', 'TH9988776', '0855555555', 'piti@example.com',       'M'),
    (2, 'สมชาย', NULL,  'ใจดี',     'Somchai', NULL,      'Jaidee',    '1990-01-15', 'B-001', '1101700200111', NULL,        '0811111111', 'somchai@example.com',    'M'),
    (2, 'วิภา',   NULL, 'สุขใจ',    'Wipa',    NULL,      'Sukjai',    '1992-09-09', 'B-002', '1101700200666', NULL,        '0866666666', 'wipa@example.com',       'F'),
    (3, 'ชูใจ',   NULL, 'เรืองรอง', 'Chujai',  NULL,      'Ruangrong', '1998-12-01', 'C-001', '1101700200777', NULL,        '0877777777', NULL,                     'F')
ON CONFLICT (hospital_id, patient_hn) DO NOTHING;
