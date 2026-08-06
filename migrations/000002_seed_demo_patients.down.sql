-- Reverses 000002_seed_demo_patients.up.sql.
--
-- Deletes by (hospital_id, patient_hn), the pair the rows were inserted
-- under, so a patient that arrived later from a real HIS is left alone.

DELETE FROM patients
WHERE (hospital_id, patient_hn) IN (
    (1, 'A-001'),
    (1, 'A-002'),
    (1, 'A-003'),
    (1, 'A-004'),
    (1, 'A-005'),
    (2, 'B-001'),
    (2, 'B-002'),
    (3, 'C-001')
);
