CREATE TRIGGER validate_insert_categories
BEFORE INSERT ON categories
FOR EACH ROW
WHEN EXISTS (SELECT 1 FROM categories WHERE priority = NEW.priority)
BEGIN
    SELECT RAISE(ABORT, 'Error: This value already exists in the table.');
END;

CREATE TRIGGER validate_update_category_priority
BEFORE UPDATE OF priority ON categories
FOR EACH ROW
WHEN EXISTS (SELECT 1 FROM categories
             WHERE priority = NEW.priority
             AND id != OLD.id) -- Ensures it doesn't block updating the same row with its current value
BEGIN
    SELECT RAISE(ABORT, 'Error: This value already exists in another row.');
END;


insert into transactions(name, amount, date, account, source, id, category_id) values('test', 323.98, '2026-02-04', 'citi', 'citi', '3b9a4061-245f-4f2c-8ba4-9c085fd7757e', 32);
insert into categories(label, priority) values('test', 1);

SELECT ticker, SUM(CASE WHEN type = 'sell' THEN -shares ELSE shares END) as shares FROM trades GROUP BY ticker;
