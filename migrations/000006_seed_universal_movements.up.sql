-- Starter library of universal movements (coach_id NULL) so a new coach can build a plan, and see
-- every muscle group and equipment filter return something, before adding movements of their own.
-- media_url stays NULL until an upload flow exists. Fixed ids let the down migration remove exactly these rows.
INSERT INTO movement (id, coach_id, name, category, muscle_group, equipment) VALUES
    ('f9bde4b6-b9e9-447b-9ad1-de0d92585b7b', NULL, 'Jumping jacks',        'warmup',   'full_body', 'bodyweight'),
    ('9b50c835-ecd0-490c-ac93-a3346fe2cd68', NULL, 'Arm circles',          'warmup',   'shoulders', 'bodyweight'),
    ('a7634543-1bbe-4c06-bc3d-018121caf9ae', NULL, 'Band pull-apart',      'warmup',   'shoulders', 'band'),
    ('0818d30a-e362-4308-9203-6451a35daea0', NULL, 'Bodyweight squat',     'warmup',   'legs',      'bodyweight'),
    ('e3935578-9941-45f9-b273-4ce0086cb0b8', NULL, 'Back squat',           'strength', 'legs',      'barbell'),
    ('a59b27f5-3cc4-45d6-9c0c-eb0a54ca996c', NULL, 'Walking lunge',        'strength', 'legs',      'dumbbell'),
    ('69dc08c5-0c75-4694-a60b-b6f2de0e5b03', NULL, 'Romanian deadlift',    'strength', 'legs',      'barbell'),
    ('be91e241-5d50-4545-b983-b5319ebc6b2e', NULL, 'Bench press',          'strength', 'chest',     'barbell'),
    ('4ff5f05d-db53-402b-8a3a-2ab07eac3edc', NULL, 'Push-up',              'strength', 'chest',     'bodyweight'),
    ('e0a40177-7338-4603-b312-73d68a0ce5f3', NULL, 'Pull-up',              'strength', 'back',      'bodyweight'),
    ('8d87cf29-1657-4148-b791-848b94767500', NULL, 'Seated cable row',     'strength', 'back',      'cable'),
    ('d7e69ef0-c665-4ce1-8920-384440f98bdb', NULL, 'Overhead press',       'strength', 'shoulders', 'barbell'),
    ('a8374134-3d10-45b3-8d78-aeed62be67c7', NULL, 'Dumbbell biceps curl', 'strength', 'arms',      'dumbbell'),
    ('ff09bfc2-d426-4ea7-99ae-eabc56070fbd', NULL, 'Triceps pushdown',     'strength', 'arms',      'cable'),
    ('dd391d27-1bab-4e53-ad17-4392515f7313', NULL, 'Plank',                'strength', 'core',      'bodyweight'),
    ('94982752-c8ae-44c6-840e-a476e33cb175', NULL, 'Kettlebell swing',     'strength', 'full_body', 'kettlebell'),
    ('59b93d7d-e948-4715-985f-589239d83a16', NULL, 'Treadmill run',        'cardio',   'legs',      'machine'),
    ('54218934-1bb8-4ddc-a629-4fde3b54806b', NULL, 'Rowing machine',       'cardio',   'full_body', 'machine'),
    ('5ed446f2-d4f0-4369-8b82-c23ed6d40986', NULL, 'Stationary bike',      'cardio',   'legs',      'machine'),
    ('0ff66926-4447-4c4e-8271-4b7e5ab2d0d1', NULL, 'Jump rope',            'cardio',   'full_body', 'other');
