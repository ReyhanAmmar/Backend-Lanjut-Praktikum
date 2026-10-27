INSERT INTO students (nim, name, grade, is_active, password, role)
SELECT 'mhs' || LPAD(g::TEXT, 3, '0'),
       'Mahasiswa ' || LPAD(g::TEXT, 3, '0'),
       80.0,
       TRUE,
       '!',
       'user'
FROM generate_series(1, 60) AS g
ON CONFLICT (nim) DO NOTHING;