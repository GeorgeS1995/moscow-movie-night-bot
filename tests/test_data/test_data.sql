INSERT INTO users(telegram_id)
VALUES (100),
       (200);


INSERT INTO movies(label, status, user_id)
VALUES ('Film1',1, (SELECT id FROM users WHERE users.telegram_id = 100)),
       ('Film2',1, (SELECT id FROM users WHERE users.telegram_id = 200)),
       ('WatchedFilm',0, (SELECT id FROM users WHERE users.telegram_id = 100))