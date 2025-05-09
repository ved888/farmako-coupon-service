BEGIN

ALTER TABLE coupon_applicable_medicines
ADD CONSTRAINT fk_medicine FOREIGN KEY (medicine_id) REFERENCES medicines(id);

ALTER TABLE coupon_applicable_categories
ADD CONSTRAINT fk_category FOREIGN KEY (category) REFERENCES categories(id);

COMMIT;