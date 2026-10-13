-- Hard delete on purpose, and it fails while any plan still uses one of these movements
-- (block_movement's foreign key has no ON DELETE), so a rollback can't strip movements out of plans.
DELETE FROM movement WHERE id IN (
    'f9bde4b6-b9e9-447b-9ad1-de0d92585b7b',
    '9b50c835-ecd0-490c-ac93-a3346fe2cd68',
    'a7634543-1bbe-4c06-bc3d-018121caf9ae',
    '0818d30a-e362-4308-9203-6451a35daea0',
    'e3935578-9941-45f9-b273-4ce0086cb0b8',
    'a59b27f5-3cc4-45d6-9c0c-eb0a54ca996c',
    '69dc08c5-0c75-4694-a60b-b6f2de0e5b03',
    'be91e241-5d50-4545-b983-b5319ebc6b2e',
    '4ff5f05d-db53-402b-8a3a-2ab07eac3edc',
    'e0a40177-7338-4603-b312-73d68a0ce5f3',
    '8d87cf29-1657-4148-b791-848b94767500',
    'd7e69ef0-c665-4ce1-8920-384440f98bdb',
    'a8374134-3d10-45b3-8d78-aeed62be67c7',
    'ff09bfc2-d426-4ea7-99ae-eabc56070fbd',
    'dd391d27-1bab-4e53-ad17-4392515f7313',
    '94982752-c8ae-44c6-840e-a476e33cb175',
    '59b93d7d-e948-4715-985f-589239d83a16',
    '54218934-1bb8-4ddc-a629-4fde3b54806b',
    '5ed446f2-d4f0-4369-8b82-c23ed6d40986',
    '0ff66926-4447-4c4e-8271-4b7e5ab2d0d1'
);
